package taggolib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/eaburns/bit"
)

const (
	// Tags specific to ID3v2 MP3
	mp3TagEncoder = "ENCODER"
	mp3TagLength  = "LENGTH"

	// Samples per frame for MPEG1 Layer III
	mp3SamplesPerFrame = 1152
)

var (
	// mp3MagicNumber is the magic number used to identify a MP3 audio stream
	mp3MagicNumber = []byte("ID3")
	// mp3APICFrame is the name of the APIC, or attached picture ID3 frame
	mp3APICFrame = []byte("APIC")
	// mp3XingMarker is the bytes which identify a Xing VBR header
	mp3XingMarker = []byte("Xing")
	// mp3InfoMarker is the bytes which identify a Info VBR header
	mp3InfoMarker = []byte("Info")
)

// mp3Parser represents a MP3 audio metadata tag parser
type mp3Parser struct {
	id3Header  *mp3ID3v2Header
	mp3Header  *mp3Header
	reader     io.ReadSeeker
	tags       map[string]string
	xingHeader *mp3XingHeader
}

// Album returns the Album tag for this stream
func (m mp3Parser) Album() string {
	return m.tags[tagAlbum]
}

// AlbumArtist returns the AlbumArtist tag for this stream
func (m mp3Parser) AlbumArtist() string {
	return m.tags[tagAlbumArtist]
}

// Artist returns the Artist tag for this stream
func (m mp3Parser) Artist() string {
	return m.tags[tagArtist]
}

// BitDepth returns the bits-per-sample of this stream
func (m mp3Parser) BitDepth() int {
	return 16
}

// Bitrate calculates the audio bitrate for this stream
func (m mp3Parser) Bitrate() int {
	// Check for a Xing header, meaning that the bitrate was calculated there
	if m.xingHeader != nil && m.xingHeader.Bitrate > 0 {
		return m.xingHeader.Bitrate
	}

	// Return bitrate from MP3 header
	return mp3BitrateMap[m.mp3Header.Bitrate]
}

// Channels returns the number of channels for this stream
func (m mp3Parser) Channels() int {
	return mp3ChannelModeMap[m.mp3Header.ChannelMode]
}

// Comment returns the Comment tag for this stream
func (m mp3Parser) Comment() string {
	return m.tags[tagComment]
}

// Date returns the Date tag for this stream
func (m mp3Parser) Date() string {
	return m.tags[tagDate]
}

// DiscNumber returns the DiscNumber tag for this stream
func (m mp3Parser) DiscNumber() int {
	disc, err := strconv.Atoi(m.tags[tagDiscNumber])
	if err != nil {
		return 0
	}

	return disc
}

// Duration returns the time duration for this stream
func (m mp3Parser) Duration() time.Duration {
	// Check for a Xing header, meaning that the duration was calculated there
	if m.xingHeader != nil && m.xingHeader.Duration > 0 {
		return time.Duration(m.xingHeader.Duration) * time.Second
	}

	// Parse length tag as integer
	length, err := strconv.Atoi(m.tags[mp3TagLength])
	if err != nil {
		return time.Duration(0 * time.Second)
	}

	return time.Duration(length/1000) * time.Second
}

// Encoder returns the encoder for this stream
func (m mp3Parser) Encoder() string {
	return m.tags[mp3TagEncoder]
}

// Format returns the name of the MP3 format
func (m mp3Parser) Format() string {
	return "MP3"
}

// Genre returns the Genre tag for this stream
func (m mp3Parser) Genre() string {
	return m.tags[tagGenre]
}

// SampleRate returns the sample rate in Hertz for this stream
func (m mp3Parser) SampleRate() int {
	return mp3SampleRateMap[m.mp3Header.SampleRate]
}

// Tag attempts to return the raw, unprocessed tag with the specified name for this stream
func (m mp3Parser) Tag(name string) string {
	return m.tags[name]
}

// Title returns the Title tag for this stream
func (m mp3Parser) Title() string {
	return m.tags[tagTitle]
}

// TrackNumber returns the TrackNumber tag for this stream
func (m mp3Parser) TrackNumber() int {
	// Check for a /, such as 2/8
	track, err := strconv.Atoi(strings.Split(m.tags[tagTrackNumber], "/")[0])
	if err != nil {
		return 0
	}

	return track
}

// newMP3Parser creates a parser for MP3 audio streams
func newMP3Parser(reader io.ReadSeeker) (*mp3Parser, error) {
	// Create MP3 parser
	parser := &mp3Parser{
		reader: reader,
	}

	// Parse ID3v2 header
	if err := parser.parseID3v2Header(); err != nil {
		return nil, err
	}

	// Parse ID3v2 frames
	if err := parser.parseID3v2Frames(); err != nil {
		return nil, err
	}

	// Parse MP3 header
	if err := parser.parseMP3Header(); err != nil {
		return nil, err
	}

	// Return parser
	return parser, nil
}

// parseID3v2Header parses the ID3v2 header at the start of an MP3 stream
func (m *mp3Parser) parseID3v2Header() error {
	// Create and use a bit reader to parse the following fields
	//   8 - ID3v2 major version
	//   8 - ID3v2 minor version
	//   1 - Unsynchronization (boolean) (ID3v2.3+)
	//   1 - Extended (boolean) (ID3v2.3+)
	//   1 - Experimental (boolean) (ID3v2.3+)
	//   1 - Footer (boolean) (ID3v2.4+)
	//   4 - (empty)
	//  32 - Size
	fields, err := bit.NewReader(m.reader).ReadFields(8, 8, 1, 1, 1, 1, 4, 32)
	if err != nil {
		return err
	}

	// Generate ID3v2 header
	m.id3Header = &mp3ID3v2Header{
		MajorVersion:      uint8(fields[0]),
		MinorVersion:      uint8(fields[1]),
		Unsynchronization: fields[2] == 1,
		Extended:          fields[3] == 1,
		Experimental:      fields[4] == 1,
		Footer:            fields[5] == 1,
		Size:              uint32(fields[7]),
	}

	// Ensure ID3v2 version is supported
	if m.id3Header.MajorVersion < 2 || m.id3Header.MajorVersion > 4 {
		return TagError{
			Err:     errUnsupportedVersion,
			Format:  m.Format(),
			Details: fmt.Sprintf("unsupported ID3 version: ID3v2.%d.%d", m.id3Header.MajorVersion, m.id3Header.MinorVersion),
		}
	}

	// Ensure Footer boolean is not defined prior to ID3v2.4
	if m.id3Header.MajorVersion < 4 && m.id3Header.Footer {
		return TagError{
			Err:     errInvalidStream,
			Format:  m.Format(),
			Details: "ID3 footer bit set prior to version ID3v2.4",
		}
	}

	// Check for extended header
	if m.id3Header.Extended {
		// Read size of extended header
		var headerSize uint32
		if err := binary.Read(m.reader, binary.BigEndian, &headerSize); err != nil {
			return err
		}

		// Seek past extended header (minus size of uint32 read), since the information
		// is irrelevant for tag parsing
		if _, err := m.reader.Seek(int64(headerSize)-4, 1); err != nil {
			return err
		}
	}

	return nil
}

// parseID3v2Frames parses ID3v2 frames from an MP3 stream
func (m *mp3Parser) parseID3v2Frames() error {
	// Store discovered tags in map
	tagMap := map[string]string{}

	// Allocate a buffer to store frame titles
	//   - ID3v2.2:  3 bytes
	//   - ID3v2.3+: 4 bytes
	var frameBuf []byte
	if m.id3Header.MajorVersion == 2 {
		frameBuf = make([]byte, 3)
	} else {
		frameBuf = make([]byte, 4)
	}

	// Create buffers for frame information
	var frameLength uint32
	tagBuf := make([]byte, 2048)
	var bufLen = uint32(len(tagBuf))

	// Byte slices which should be trimmed and discarded from prefix or suffix
	trimPrefix := []byte{255, 254}
	trimSuffix := []byte{0}

	// Continuously loop and parse frames
	for {
		// Parse a frame title
		if _, err := m.reader.Read(frameBuf); err != nil {
			return err
		}

		// Stop parsing frames when frame title is nil, because we have reached padding
		if frameBuf[0] == byte(0) {
			break
		}

		// If byte 255 discovered, we have reached the start of the MP3 header
		if frameBuf[0] == byte(255) {
			// Read in more bytes to enable fetching the Xing header
			if _, err := m.reader.Read(tagBuf); err != nil {
				return err
			}

			// Pre-seed the current data as a bytes reader, to parse MP3 header, while also
			// appending more bytes to find the Xing header
			m.reader = bytes.NewReader(append(frameBuf, tagBuf...))
			break
		}

		// Parse the length of the frame data
		//   - ID3v2.2:  24-bit integer, big endian
		//   - ID3v2.3+: 32-bit integer, big endian
		if m.id3Header.MajorVersion == 2 {
			// Read 3 bytes to parse length
			if _, err := m.reader.Read(tagBuf[:3]); err != nil {
				return err
			}

			// Store frame length
			// Thanks: https://github.com/ascherkus/go-id3/blob/master/src/id3/id3v22.go#L24
			frameLength = uint32(tagBuf[0])<<16 | uint32(tagBuf[1])<<8 | uint32(tagBuf[2])
		} else {
			// Read 4 bytes as uint32 to parse length
			if err := binary.Read(m.reader, binary.BigEndian, &frameLength); err != nil {
				return err
			}

			// ID3v2.3+: Skip over frame flags
			if _, err := m.reader.Seek(2, 1); err != nil {
				return err
			}
		}

		// If frame is attached picture OR frame is too long for buffer, seek past it
		if bytes.Equal(frameBuf, mp3APICFrame) || frameLength > bufLen {
			// Seek past picture data and continue loop
			if _, err := m.reader.Seek(int64(frameLength), 1); err != nil {
				return err
			}

			continue
		}

		// Parse the frame data tag
		n, err := m.reader.Read(tagBuf[:frameLength])
		if err != nil {
			return err
		}

		// Trim leading bytes such as UTF-8 BOM, garbage bytes, trim trailing nil
		// BUG(mdlayher): MP3: handle ID3 tag encodings that aren't UTF-8, stored in tagBuf[0]
		tag := string(bytes.TrimPrefix(bytes.TrimSuffix(tagBuf[1:n], trimSuffix), trimPrefix))

		// Map frame title to tag title, store frame data
		tagMap[mp3ID3v2FrameToTag[string(frameBuf)]] = tag
	}

	// Store tags in parser
	m.tags = tagMap
	return nil
}

// mp3ID3v2FrameToTag maps a MP3 ID3v2 frame title to its actual tag name
var mp3ID3v2FrameToTag = map[string]string{
	// ID3v2.2
	"TAL": tagAlbum,
	"TRK": tagTrackNumber,
	"TP1": tagArtist,
	"TP2": tagAlbumArtist,
	"TT2": tagTitle,
	"TYE": tagDate,
	"TPA": tagDiscNumber,
	"TCO": tagGenre,

	// ID3v2.3+
	"COMM": tagComment,
	"TALB": tagAlbum,
	"TCON": tagGenre,
	"TDRC": tagDate,
	"TIT2": tagTitle,
	"TLEN": mp3TagLength,
	"TPE1": tagArtist,
	"TPE2": tagAlbumArtist,
	"TPOS": tagDiscNumber,
	"TRCK": tagTrackNumber,
	"TSSE": mp3TagEncoder,
	"TYER": tagDate,
}

// mp3ID3v2Header represents the MP3 ID3v2 header section
type mp3ID3v2Header struct {
	MajorVersion      uint8
	MinorVersion      uint8
	Unsynchronization bool
	Extended          bool
	Experimental      bool
	Footer            bool
	Size              uint32
}

// mp3ID3v2ExtendedHeader reperesents the optional MP3 ID3v2 extended header section
type mp3ID3v2ExtendedHeader struct {
	HeaderSize   uint32
	CRC32Present bool
	PaddingSize  uint32
}

// parseMP3Header parses the MP3 header after the ID3 headers in a MP3 stream
func (m *mp3Parser) parseMP3Header() error {
	// Read buffers continuously until we reach end of padding section, and find the
	// MP3 header, which starts with byte 255
	headerBuf := make([]byte, 4096)
	for {
		if _, err := m.reader.Read(headerBuf); err != nil {
			return err
		}

		// If first byte is 255, value was pre-seeded by tag parser
		if headerBuf[0] == byte(255) {
			break
		}

		// Search for byte 255
		index := bytes.Index(headerBuf, []byte{255})
		if index != -1 {
			// We have encountered the header, re-slice forward to its index, and read 64 more
			// bytes to ensure that the Xing header is retrieved
			tempBuf := make([]byte, 64)
			if _, err := m.reader.Read(tempBuf); err != nil {
				return err
			}

			// Append buffers to add Xing header
			headerBuf = append(headerBuf[index:], tempBuf...)
			break
		}
	}

	// Create and use a bit reader to parse the following fields
	//  11 - MP3 frame sync (all bits set)
	//   2 - MPEG audio version ID
	//   2 - Layer description
	//   1 - Protection bit (boolean)
	//   4 - Bitrate index
	//   2 - Sample rate index
	//   1 - Padding (boolean)
	//   1 - Private (boolean)
	//   2 - Channel mode index
	//   2 - Mode extension
	//   1 - Copyright (boolean)
	//   1 - Original (boolean)
	//   2 - Emphasis
	fields, err := bit.NewReader(bytes.NewReader(headerBuf)).ReadFields(11, 2, 2, 1, 4, 2, 1, 1, 2, 2, 1, 1, 2)
	if err != nil {
		return err
	}

	// Create output MP3 header
	m.mp3Header = &mp3Header{
		MPEGVersionID: uint8(fields[1]),
		MPEGLayerID:   uint8(fields[2]),
		Protected:     fields[3] == 0,
		Bitrate:       uint16(fields[4]),
		SampleRate:    uint16(fields[5]),
		Padding:       fields[6] == 1,
		Private:       fields[7] == 1,
		ChannelMode:   uint8(fields[8]),
		ModeExtension: uint8(fields[9]),
		Copyright:     fields[10] == 1,
		Original:      fields[11] == 1,
		Emphasis:      uint8(fields[12]),
	}

	// Check to make sure we are parsing MPEG Version 1, Layer 3
	// Note: this check is correct, as these values actually map to:
	//   - Version ID 3 -> MPEG Version 1
	//   - Layer ID 1 -> MPEG Layer 3
	if m.mp3Header.MPEGVersionID != 3 {
		return TagError{
			Err:     errUnsupportedVersion,
			Format:  m.Format(),
			Details: fmt.Sprintf("unsupported MPEG version ID: %d", m.mp3Header.MPEGVersionID),
		}
	}

	if m.mp3Header.MPEGLayerID != 1 {
		return TagError{
			Err:     errUnsupportedVersion,
			Format:  m.Format(),
			Details: fmt.Sprintf("unsupported MPEG layer ID: %d", m.mp3Header.MPEGLayerID),
		}
	}

	// Search for "Xing" header, to help calculate duration
	index := bytes.Index(headerBuf, mp3XingMarker)
	if index == -1 {
		// Search for "Info" header, which may also be present
		index = bytes.Index(headerBuf, mp3InfoMarker)
		if index == -1 {
			// No Xing or Info header, must calculate duration via LENGTH tag
			// BUG(mdlayher): MP3: Duration of CBR files must be calculated by finding last MP3 frame header
			return nil
		}
	}

	// Re-slice forward and begin reading data we want from the Xing header, skipping
	// over the flags to directly read data
	headerBuf = headerBuf[index+len(mp3XingMarker)+4:]
	m.xingHeader = &mp3XingHeader{
		FrameCount: binary.BigEndian.Uint32(headerBuf[0:4]),
		StreamSize: binary.BigEndian.Uint32(headerBuf[4:8]),
	}

	// Calculate file duration and VBR bitrate using Xing/Info header data
	// Thanks: https://github.com/taglib/taglib/blob/master/taglib/mpeg/mpegproperties.cpp#L212
	m.xingHeader.Duration = int((float64(mp3SamplesPerFrame) / float64(m.SampleRate())) * float64(m.xingHeader.FrameCount))
	m.xingHeader.Bitrate = int(float64(m.xingHeader.StreamSize*8) / float64(m.xingHeader.Duration) / 1000)

	// If bitrate calculated is above 320, correct it to 320, per specification
	if m.xingHeader.Bitrate > 320 {
		m.xingHeader.Bitrate = 320
	}

	return nil
}

// mp3XingHeader represents additional information contained within a Xing header, used to
// help parse MP3 duration
type mp3XingHeader struct {
	FrameCount uint32
	StreamSize uint32
	Duration   int
	Bitrate    int
}

// mp3Header represents a MP3 audio stream header, and contains information about the stream
type mp3Header struct {
	MPEGVersionID uint8
	MPEGLayerID   uint8
	Protected     bool
	Bitrate       uint16
	SampleRate    uint16
	Padding       bool
	Private       bool
	ChannelMode   uint8
	ModeExtension uint8
	Copyright     bool
	Original      bool
	Emphasis      uint8
}

// mp3BitrateMap maps MPEG Layer 3 Version 1 bitrate to its actual rate
var mp3BitrateMap = map[uint16]int{
	0:  0,
	1:  32,
	2:  40,
	3:  48,
	4:  56,
	5:  64,
	6:  80,
	7:  96,
	8:  112,
	9:  128,
	10: 160,
	11: 192,
	12: 224,
	13: 256,
	14: 320,
}

// mp3SampleRateMap maps MPEG Layer 3 Version 1 sample rate to its actual rate
var mp3SampleRateMap = map[uint16]int{
	0: 44100,
	1: 48000,
	2: 32000,
}

// mp3ChannelModeMap maps MPEG Layer 3 Version 1 channels to the number of channels
var mp3ChannelModeMap = map[uint8]int{
	0: 2,
	1: 2,
	3: 2,
	4: 1,
}

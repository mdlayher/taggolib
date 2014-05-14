package taggolib

import (
	"encoding/binary"
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
)

var (
	// mp3MagicNumber is the magic number used to identify a MP3 audio stream
	mp3MagicNumber = []byte("ID3")
)

// MP3Parser represents a MP3 audio metadata tag parser
type MP3Parser struct {
	id3Header *mp3ID3v2Header
	mp3Header *mp3Header
	reader    io.ReadSeeker
	tags      map[string]string
}

// Album returns the Album tag for this stream
func (m MP3Parser) Album() string {
	return m.tags[tagAlbum]
}

// AlbumArtist returns the AlbumArtist tag for this stream
func (m MP3Parser) AlbumArtist() string {
	return m.tags[tagAlbumArtist]
}

// Artist returns the Artist tag for this stream
func (m MP3Parser) Artist() string {
	return m.tags[tagArtist]
}

// BitDepth returns the bits-per-sample of this stream
func (m MP3Parser) BitDepth() int {
	return 16
}

// Bitrate calculates the audio bitrate for this stream
func (m MP3Parser) Bitrate() int {
	return mp3BitrateMap[m.mp3Header.Bitrate]
}

// Channels returns the number of channels for this stream
func (m MP3Parser) Channels() int {
	return mp3ChannelModeMap[m.mp3Header.ChannelMode]
}

// Checksum returns the checksum for this stream
func (m MP3Parser) Checksum() string {
	return ""
}

// Comment returns the Comment tag for this stream
func (m MP3Parser) Comment() string {
	return m.tags[tagComment]
}

// Date returns the Date tag for this stream
func (m MP3Parser) Date() string {
	return m.tags[tagDate]
}

// Duration returns the time duration for this stream
func (m MP3Parser) Duration() time.Duration {
	// Parse length as integer
	length, err := strconv.Atoi(m.tags[mp3TagLength])
	if err != nil {
		return time.Duration(0 * time.Second)
	}

	return time.Duration(length/1000) * time.Second
}

// Encoder returns the encoder for this stream
func (m MP3Parser) Encoder() string {
	return m.tags[mp3TagEncoder]
}

// Format returns the name of the MP3 format
func (m MP3Parser) Format() string {
	return "MP3"
}

// Genre returns the Genre tag for this stream
func (m MP3Parser) Genre() string {
	return m.tags[tagGenre]
}

// SampleRate returns the sample rate in Hertz for this stream
func (m MP3Parser) SampleRate() int {
	return mp3SampleRateMap[m.mp3Header.SampleRate]
}

// Tag attempts to return the raw, unprocessed tag with the specified name for this stream
func (m MP3Parser) Tag(name string) string {
	return m.tags[name]
}

// Title returns the Title tag for this stream
func (m MP3Parser) Title() string {
	return m.tags[tagTitle]
}

// TrackNumber returns the TrackNumber tag for this stream
func (m MP3Parser) TrackNumber() int {
	// Check for a /, such as 2/8
	track, err := strconv.Atoi(strings.Split(m.tags[tagTrackNumber], "/")[0])
	if err != nil {
		return 0
	}

	return track
}

// newMP3Parser creates a parser for MP3 audio streams
func newMP3Parser(reader io.ReadSeeker) (*MP3Parser, error) {
	// Create MP3 parser
	parser := &MP3Parser{
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
func (m *MP3Parser) parseID3v2Header() error {
	// Create and use a bit reader to parse the following fields
	//   8 - ID3v2 major version
	//   8 - ID3v2 minor version
	//   1 - Unsynchronization (boolean)
	//   1 - Extended (boolean)
	//   1 - Experimental (boolean)
	//   1 - Footer (boolean)
	//   4 - (empty)
	//  24 - Size
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

	// Check for extended header
	if m.id3Header.Extended {
		// Read size of extended header
		var headerSize uint32
		if err := binary.Read(m.reader, binary.BigEndian, &headerSize); err != nil {
			return err
		}

		// Seek past extended header (minus size of uint32 read), since the information
		// is irrelevant for tag parsing
		if _, err := m.reader.Seek(int64(headerSize) - 4, 1); err != nil {
			return err
		}
	}

	return nil
}

// parseID3v2Frames parses ID3v2 frames from an MP3 stream
func (m *MP3Parser) parseID3v2Frames() error {
	// Continuously loop and parse frames
	tagMap := map[string]string{}
	for {
		// Parse a frame title
		frameBuf := make([]byte, 4)
		if _, err := m.reader.Read(frameBuf); err != nil {
			return err
		}

		// Stop parsing frames when frame title is nil
		if frameBuf[0] == byte(0) {
			// Seek 8 bytes ahead to MP3 audio stream
			if _, err := m.reader.Seek(8, 1); err != nil {
				return err
			}

			// Break tag parsing loop
			break
		}

		// Parse the length of the frame data
		var frameLength uint32
		if err := binary.Read(m.reader, binary.BigEndian, &frameLength); err != nil {
			return err
		}

		// Skip over frame flags
		if _, err := m.reader.Seek(2, 1); err != nil {
			return err
		}

		// Parse the frame data tag
		tagBuf := make([]byte, frameLength)
		if _, err := m.reader.Read(tagBuf); err != nil {
			return err
		}

		// Map frame title to tag title, store frame data, stripping UTF-8 BOM
		tagMap[mp3ID3v2FrameToTag[string(frameBuf)]] = string(tagBuf[1:])
	}

	// Store tags in parser
	m.tags = tagMap
	return nil
}

// mp3ID3v2FrameToTag maps a MP3 ID3v2 frame title to its actual tag name
var mp3ID3v2FrameToTag = map[string]string{
	"COMM": tagComment,
	"TALB": tagAlbum,
	"TCON": tagGenre,
	"TDRC": tagDate,
	"TIT2": tagTitle,
	"TLEN": mp3TagLength,
	"TPE1": tagArtist,
	"TPE2": tagAlbumArtist,
	"TRCK": tagTrackNumber,
	"TSSE": mp3TagEncoder,
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
func (m *MP3Parser) parseMP3Header() error {
	// Create and use a bit reader to parse the following fields
	//  11 - MP3 frame sync (all bits set)
	//   2 - MPEG audio version ID
	//   2 - Layer description
	//   1 - Protection bit (boolean)
	//   4 - Bitrate index
	fields, err := bit.NewReader(m.reader).ReadFields(11, 2, 2, 1, 4, 2, 1, 1, 2)
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
	}

	return nil
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

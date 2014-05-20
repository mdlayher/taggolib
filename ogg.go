package taggolib

import (
	"bytes"
	"encoding/binary"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/eaburns/bit"
)

var (
	// oggMagicNumber is the magic number used to identify a OGG audio stream
	oggMagicNumber = []byte("OggS")
	// oggVorbisWord is used to denote the beginning of a Vorbis information block
	oggVorbisWord = []byte("vorbis")
)

// OGGParser represents a OGG audio metadata tag parser
type OGGParser struct {
	buffer   []byte
	encoder  string
	idHeader *oggIDHeader
	reader   io.ReadSeeker
	tags     map[string]string
	ui32     uint32
}

// Album returns the Album tag for this stream
func (o OGGParser) Album() string {
	return o.tags[tagAlbum]
}

// AlbumArtist returns the AlbumArtist tag for this stream
func (o OGGParser) AlbumArtist() string {
	return o.tags[tagAlbumArtist]
}

// Artist returns the Artist tag for this stream
func (o OGGParser) Artist() string {
	return o.tags[tagArtist]
}

// BitDepth returns the bits-per-sample of this stream
func (o OGGParser) BitDepth() int {
	// Ogg Vorbis should always provide 16 bit depth
	return 16
}

// Bitrate calculates the audio bitrate for this stream
func (o OGGParser) Bitrate() int {
	// TODO: see how max/min bitrate play into calculations
	return int(o.idHeader.NomBitrate) / 1000
}

// Channels returns the number of channels for this stream
func (o OGGParser) Channels() int {
	return int(o.idHeader.ChannelCount)
}

// Comment returns the Comment tag for this stream
func (o OGGParser) Comment() string {
	return o.tags[tagComment]
}

// Date returns the Date tag for this stream
func (o OGGParser) Date() string {
	return o.tags[tagDate]
}

// DiscNumber returns the DiscNumber tag for this stream
func (o OGGParser) DiscNumber() int {
	disc, err := strconv.Atoi(o.tags[tagDiscNumber])
	if err != nil {
		return 0
	}

	return disc
}

// Duration returns the time duration for this stream
func (o OGGParser) Duration() time.Duration {
	// TODO: calculate this using last granular position detected, or find a better method
	return time.Duration(0 * time.Second)
}

// Encoder returns the encoder for this stream
func (o OGGParser) Encoder() string {
	return o.encoder
}

// Format returns the name of the OGG format
func (o OGGParser) Format() string {
	return "OGG"
}

// Genre returns the Genre tag for this stream
func (o OGGParser) Genre() string {
	return o.tags[tagGenre]
}

// SampleRate returns the sample rate in Hertz for this stream
func (o OGGParser) SampleRate() int {
	return int(o.idHeader.SampleRate)
}

// Tag attempts to return the raw, unprocessed tag with the specified name for this stream
func (o OGGParser) Tag(name string) string {
	return o.tags[name]
}

// Title returns the Title tag for this stream
func (o OGGParser) Title() string {
	return o.tags[tagTitle]
}

// TrackNumber returns the TrackNumber tag for this stream
func (o OGGParser) TrackNumber() int {
	// Check for a /, such as 2/8
	track, err := strconv.Atoi(strings.Split(o.tags[tagTrackNumber], "/")[0])
	if err != nil {
		return 0
	}

	return track
}

// newOGGParser creates a parser for OGG audio streams
func newOGGParser(reader io.ReadSeeker) (*OGGParser, error) {
	// Create OGG parser
	parser := &OGGParser{
		buffer: make([]byte, 128),
		reader: reader,
	}

	// Parse the required ID header
	if err := parser.parseOGGIDHeader(); err != nil {
		return nil, err
	}

	// Parse the required comment header
	if err := parser.parseOGGCommentHeader(); err != nil {
		return nil, err
	}

	// Parse the required setup header
	if err := parser.parseOGGSetupHeader(); err != nil {
		return nil, err
	}

	// Return parser
	return parser, nil
}

// oggPageHeader represents the information contained in an Ogg Page header
type oggPageHeader struct {
	CapturePattern  []byte
	Version         uint8
	HeaderType      uint8
	GranulePosition uint64
	BitstreamSerial uint32
	PageSequence    uint32
	Checksum        []byte
	PageSegments    uint8
}

// parseOGGPageHeader parses an Ogg page header
func (o *OGGParser) parseOGGPageHeader(skipMagicNumber bool) (*oggPageHeader, error) {
	// Create page header
	pageHeader := new(oggPageHeader)

	// Unless skip is specified, check for capture pattern
	if !skipMagicNumber {
		if _, err := o.reader.Read(o.buffer[:4]); err != nil {
			return nil, err
		}
		pageHeader.CapturePattern = o.buffer[:4]

		// Verify proper capture pattern
		if !bytes.Equal(pageHeader.CapturePattern, oggMagicNumber) {
			return nil, ErrInvalidStream
		}
	} else {
		// If skipped, assume capture pattern is correct magic number
		pageHeader.CapturePattern = oggMagicNumber
	}

	// Read other bits in an Ogg Page
	var uint8Buf uint8
	var uint64Buf uint64

	// Version (must always be 0)
	if err := binary.Read(o.reader, binary.LittleEndian, &uint8Buf); err != nil {
		return nil, err
	}
	pageHeader.Version = uint8Buf

	// Verify mandated version 0
	if pageHeader.Version != 0 {
		return nil, ErrInvalidStream
	}

	// Header type
	if err := binary.Read(o.reader, binary.LittleEndian, &uint8Buf); err != nil {
		return nil, err
	}
	pageHeader.HeaderType = uint8Buf

	// Granule position
	if err := binary.Read(o.reader, binary.LittleEndian, &uint64Buf); err != nil {
		return nil, err
	}
	pageHeader.GranulePosition = uint64Buf

	// Bitstream serial number
	if err := binary.Read(o.reader, binary.LittleEndian, &o.ui32); err != nil {
		return nil, err
	}
	pageHeader.BitstreamSerial = o.ui32

	// Page sequence number
	if err := binary.Read(o.reader, binary.LittleEndian, &o.ui32); err != nil {
		return nil, err
	}
	pageHeader.PageSequence = o.ui32

	// Checksum
	if _, err := o.reader.Read(o.buffer[:4]); err != nil {
		return nil, err
	}
	pageHeader.Checksum = o.buffer[:4]

	// Page segments
	if err := binary.Read(o.reader, binary.LittleEndian, &uint8Buf); err != nil {
		return nil, err
	}
	pageHeader.PageSegments = uint8Buf

	// Segment table is next, but we won't need it for tag parsing, so seek ahead
	// size of uint8 (1 byte) multiplied by number of page segments
	if _, err := o.reader.Seek(int64(pageHeader.PageSegments), 1); err != nil {
		return nil, err

	}
	return pageHeader, nil
}

// parseOGGCommonHeader parses information common to all Ogg Vorbis headers
func (o *OGGParser) parseOGGCommonHeader() (byte, error) {
	// Read the first byte to get header type
	if _, err := o.reader.Read(o.buffer[:1]); err != nil {
		return 0, err
	}

	// Store first byte at end of buffer so we can return it later without more allocations
	o.buffer[len(o.buffer)-1] = o.buffer[0]

	// Read for 'vorbis' identification word
	if _, err := o.reader.Read(o.buffer[:len(oggVorbisWord)]); err != nil {
		return 0, err
	}

	// Ensure 'vorbis' identification word is present
	if !bytes.Equal(o.buffer[:len(oggVorbisWord)], oggVorbisWord) {
		return 0, ErrInvalidStream
	}

	// Return header type from end of buffer
	return o.buffer[len(o.buffer)-1], nil
}

// oggIDHeader represents the information contained in an Ogg Vorbis identification header
type oggIDHeader struct {
	VorbisVersion uint32
	ChannelCount  uint8
	SampleRate    uint32
	MaxBitrate    uint32
	NomBitrate    uint32
	MinBitrate    uint32
	Blocksize0    uint8
	Blocksize1    uint8
	Framing       bool
}

// parseOGGIDHeader parses the required identification header for an Ogg Vorbis stream
func (o *OGGParser) parseOGGIDHeader() error {
	// Read OGG page header, skipping the capture pattern because New() already verified
	// the magic number for us
	if _, err := o.parseOGGPageHeader(true); err != nil {
		return err
	}

	// Check for valid common header
	headerType, err := o.parseOGGCommonHeader()
	if err != nil {
		return err
	}

	// Ensure header type 1: identification header
	if headerType != byte(1) {
		return ErrInvalidStream
	}

	// Read fields found in identification header
	header := new(oggIDHeader)
	var int8Buf uint8

	// Vorbis version
	if err := binary.Read(o.reader, binary.LittleEndian, &o.ui32); err != nil {
		return err
	}
	header.VorbisVersion = o.ui32

	// Ensure Vorbis version is 0, per specification
	if header.VorbisVersion != 0 {
		return ErrInvalidStream
	}

	// Channel count
	if err := binary.Read(o.reader, binary.LittleEndian, &int8Buf); err != nil {
		return err
	}
	header.ChannelCount = int8Buf

	// uint32 x 4: sample rate, maximum bitrate, nominal bitrate, minimum bitrate
	uint32Slice := make([]uint32, 4)
	for i := 0; i < 4; i++ {
		// Read in one uint32
		if err := binary.Read(o.reader, binary.LittleEndian, &uint32Slice[i]); err != nil {
			return err
		}
	}

	// Copy out slice values
	header.SampleRate = uint32Slice[0]
	header.MaxBitrate = uint32Slice[1]
	header.NomBitrate = uint32Slice[2]
	header.MinBitrate = uint32Slice[3]

	// Create and use a bit reader to parse the following fields
	//    4 - Blocksize 0
	//    4 - Blocksize 1
	//    1 - Framing flag
	fields, err := bit.NewReader(o.reader).ReadFields(4, 4, 1)
	if err != nil {
		return err
	}

	header.Blocksize0 = uint8(fields[0])
	header.Blocksize1 = uint8(fields[1])
	header.Framing = fields[2] == 1

	// Store ID header
	o.idHeader = header
	return nil
}

// parseOGGCommentHeader parses the Vorbis Comment tags in an Ogg Vorbis file
func (o *OGGParser) parseOGGCommentHeader() error {
	// Read OGG page header, specifying false to check the capture pattern
	if _, err := o.parseOGGPageHeader(false); err != nil {
		return err
	}

	// Parse common header
	headerType, err := o.parseOGGCommonHeader()
	if err != nil {
		return err
	}

	// Verify header type (3: Vorbis Comment)
	if headerType != byte(3) {
		return ErrInvalidStream
	}

	// Read vendor length
	if err := binary.Read(o.reader, binary.LittleEndian, &o.ui32); err != nil {
		return err
	}

	// Read vendor string, store as encoder
	if _, err := o.reader.Read(o.buffer[:o.ui32]); err != nil {
		return err
	}
	o.encoder = string(o.buffer[:o.ui32])

	// Read comment length (new allocation for use with loop counter)
	var commentLength uint32
	if err := binary.Read(o.reader, binary.LittleEndian, &commentLength); err != nil {
		return err
	}

	// Begin iterating tags, and building tag map
	tagMap := map[string]string{}
	for i := 0; i < int(commentLength); i++ {
		// Read tag string length
		if err := binary.Read(o.reader, binary.LittleEndian, &o.ui32); err != nil {
			return err
		}

		// Read tag string
		n, err := o.reader.Read(o.buffer[:o.ui32])
		if err != nil {
			return err
		}

		// Split tag name and data, store in map
		pair := strings.Split(string(o.buffer[:n]), "=")
		tagMap[strings.ToUpper(pair[0])] = pair[1]
	}

	// Seek one byte forward to prepare for the setup header
	if _, err := o.reader.Seek(1, 1); err != nil {
		return err
	}

	// Store tags
	o.tags = tagMap
	return nil
}

// parseOGGSetupHeader parses the Vorbis setup header in an Ogg Vorbis file
func (o *OGGParser) parseOGGSetupHeader() error {
	// Parse common header
	headerType, err := o.parseOGGCommonHeader()
	if err != nil {
		return err
	}

	// Verify header type (5: Vorbis Setup)
	if headerType != byte(5) {
		return ErrInvalidStream
	}

	// TODO: continue parsing here so we can eventually calculate duration

	return nil
}
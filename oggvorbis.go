package taggolib

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/eaburns/bit"
)

var (
	// oggMagicNumber is the magic number used to identify an OGG container audio stream
	oggMagicNumber = []byte("OggS")
	// oggVorbisVorbisWord is used to denote the beginning of a Vorbis information block
	oggVorbisVorbisWord = []byte("vorbis")
)

// oggVorbisParser represents a OGGVorbis audio metadata tag parser
type oggVorbisParser struct {
	duration time.Duration
	encoder  string
	idHeader *oggVorbisIDHeader
	reader   io.ReadSeeker
	tags     map[string]string

	// Shared buffer and unsigned integers stored as fields to prevent unneeded allocations
	buffer []byte
	ui8    uint8
	ui32   uint32
	ui64   uint64
}

// Album returns the Album tag for this stream
func (o oggVorbisParser) Album() string {
	return o.tags[tagAlbum]
}

// AlbumArtist returns the AlbumArtist tag for this stream
func (o oggVorbisParser) AlbumArtist() string {
	return o.tags[tagAlbumArtist]
}

// Artist returns the Artist tag for this stream
func (o oggVorbisParser) Artist() string {
	return o.tags[tagArtist]
}

// BitDepth returns the bits-per-sample of this stream
func (o oggVorbisParser) BitDepth() int {
	// Ogg Vorbis should always provide 16 bit depth
	return 16
}

// Bitrate calculates the audio bitrate for this stream
func (o oggVorbisParser) Bitrate() int {
	// BUG(mdlayher): Ogg Vorbis: check if maximum/minimum bitrate from headers should be used in calculation
	return int(o.idHeader.NomBitrate) / 1000
}

// Channels returns the number of channels for this stream
func (o oggVorbisParser) Channels() int {
	return int(o.idHeader.ChannelCount)
}

// Comment returns the Comment tag for this stream
func (o oggVorbisParser) Comment() string {
	return o.tags[tagComment]
}

// Date returns the Date tag for this stream
func (o oggVorbisParser) Date() string {
	return o.tags[tagDate]
}

// DiscNumber returns the DiscNumber tag for this stream
func (o oggVorbisParser) DiscNumber() int {
	disc, err := strconv.Atoi(o.tags[tagDiscNumber])
	if err != nil {
		return 0
	}

	return disc
}

// Duration returns the time duration for this stream
func (o oggVorbisParser) Duration() time.Duration {
	return o.duration
}

// Encoder returns the encoder for this stream
func (o oggVorbisParser) Encoder() string {
	return o.encoder
}

// Format returns the name of the Ogg Vorbis format
func (o oggVorbisParser) Format() string {
	return "Ogg Vorbis"
}

// Genre returns the Genre tag for this stream
func (o oggVorbisParser) Genre() string {
	return o.tags[tagGenre]
}

// Publisher returns the Publisher (record-label) tag for this stream
func (o oggVorbisParser) Publisher() string {
	return o.tags[tagPublisher]
}

// SampleRate returns the sample rate in Hertz for this stream
func (o oggVorbisParser) SampleRate() int {
	return int(o.idHeader.SampleRate)
}

// Tag attempts to return the raw, unprocessed tag with the specified name for this stream
func (o oggVorbisParser) Tag(name string) string {
	return o.tags[name]
}

// Title returns the Title tag for this stream
func (o oggVorbisParser) Title() string {
	return o.tags[tagTitle]
}

// TrackNumber returns the TrackNumber tag for this stream
func (o oggVorbisParser) TrackNumber() int {
	// Check for a /, such as 2/8
	track, err := strconv.Atoi(strings.Split(o.tags[tagTrackNumber], "/")[0])
	if err != nil {
		return 0
	}

	return track
}

// newOGGVorbisParser creates a parser for OGGVorbis audio streams
func newOGGVorbisParser(reader io.ReadSeeker) (*oggVorbisParser, error) {
	// Create OGGVorbis parser
	parser := &oggVorbisParser{
		buffer: make([]byte, 128),
		reader: reader,
	}

	// Parse the required ID header
	if err := parser.parseOGGVorbisIDHeader(); err != nil {
		return nil, err
	}

	// Parse the required comment header
	if err := parser.parseOGGVorbisCommentHeader(); err != nil {
		return nil, err
	}

	// Parse the file's duration
	if err := parser.parseOGGVorbisDuration(); err != nil {
		return nil, err
	}

	// Return parser
	return parser, nil
}

// oggVorbisPageHeader represents the information contained in an Ogg Page header
type oggVorbisPageHeader struct {
	CapturePattern  []byte
	Version         uint8
	HeaderType      uint8
	GranulePosition uint64
	BitstreamSerial uint32
	PageSequence    uint32
	Checksum        []byte
	PageSegments    uint8
}

// parseOGGVorbisPageHeader parses an Ogg page header
func (o *oggVorbisParser) parseOGGVorbisPageHeader(skipMagicNumber bool) (*oggVorbisPageHeader, error) {
	// Create page header
	pageHeader := new(oggVorbisPageHeader)

	// Unless skip is specified, check for capture pattern
	if !skipMagicNumber {
		if _, err := o.reader.Read(o.buffer[:4]); err != nil {
			return nil, err
		}
		pageHeader.CapturePattern = o.buffer[:4]

		// Verify proper capture pattern
		if !bytes.Equal(pageHeader.CapturePattern, oggMagicNumber) {
			return nil, TagError{
				Err:     errInvalidStream,
				Format:  o.Format(),
				Details: "unrecognized capture pattern in Ogg page header",
			}
		}
	} else {
		// If skipped, assume capture pattern is correct magic number
		pageHeader.CapturePattern = oggMagicNumber
	}

	// Version (must always be 0)
	if err := binary.Read(o.reader, binary.LittleEndian, &o.ui8); err != nil {
		return nil, err
	}
	pageHeader.Version = o.ui8

	// Verify mandated version 0
	if pageHeader.Version != 0 {
		return nil, TagError{
			Err:     errInvalidStream,
			Format:  o.Format(),
			Details: fmt.Sprintf("Vorbis version must be 0, but found version %d", pageHeader.Version),
		}
	}

	// Header type
	if err := binary.Read(o.reader, binary.LittleEndian, &o.ui8); err != nil {
		return nil, err
	}
	pageHeader.HeaderType = o.ui8

	// Granule position
	if err := binary.Read(o.reader, binary.LittleEndian, &o.ui64); err != nil {
		return nil, err
	}
	pageHeader.GranulePosition = o.ui64

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
	if err := binary.Read(o.reader, binary.LittleEndian, &o.ui8); err != nil {
		return nil, err
	}
	pageHeader.PageSegments = o.ui8

	// Segment table is next, but we won't need it for tag parsing, so seek ahead
	// size of uint8 (1 byte) multiplied by number of page segments
	if _, err := o.reader.Seek(int64(pageHeader.PageSegments), 1); err != nil {
		return nil, err

	}
	return pageHeader, nil
}

// parseOGGVorbisCommonHeader parses information common to all Ogg Vorbis headers
func (o *oggVorbisParser) parseOGGVorbisCommonHeader() (byte, error) {
	// Read the first byte to get header type
	if _, err := o.reader.Read(o.buffer[:1]); err != nil {
		return 0, err
	}

	// Store first byte at end of buffer so we can return it later without more allocations
	o.buffer[len(o.buffer)-1] = o.buffer[0]

	// Read for 'vorbis' identification word
	if _, err := o.reader.Read(o.buffer[:len(oggVorbisVorbisWord)]); err != nil {
		return 0, err
	}

	// Ensure 'vorbis' identification word is present
	if !bytes.Equal(o.buffer[:len(oggVorbisVorbisWord)], oggVorbisVorbisWord) {
		return 0, TagError{
			Err:     errInvalidStream,
			Format:  o.Format(),
			Details: "unrecognized identification word in header",
		}
	}

	// Return header type from end of buffer
	return o.buffer[len(o.buffer)-1], nil
}

// oggVorbisIDHeader represents the information contained in an Ogg Vorbis identification header
type oggVorbisIDHeader struct {
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

// parseOGGVorbisIDHeader parses the required identification header for an Ogg Vorbis stream
func (o *oggVorbisParser) parseOGGVorbisIDHeader() error {
	// Read OGGVorbis page header, skipping the capture pattern because New() already verified
	// the magic number for us
	if _, err := o.parseOGGVorbisPageHeader(true); err != nil {
		return err
	}

	// Check for valid common header
	headerType, err := o.parseOGGVorbisCommonHeader()
	if err != nil {
		return err
	}

	// Ensure header type 1: identification header
	if headerType != byte(1) {
		return TagError{
			Err:     errInvalidStream,
			Format:  o.Format(),
			Details: "invalid header type for identification header",
		}
	}

	// Read fields found in identification header
	header := new(oggVorbisIDHeader)

	// Vorbis version
	if err := binary.Read(o.reader, binary.LittleEndian, &o.ui32); err != nil {
		return err
	}
	header.VorbisVersion = o.ui32

	// Ensure Vorbis version is 0, per specification
	if header.VorbisVersion != 0 {
		return TagError{
			Err:     errInvalidStream,
			Format:  o.Format(),
			Details: fmt.Sprintf("Vorbis version must be 0, but found version %d", header.VorbisVersion),
		}
	}

	// Channel count
	if err := binary.Read(o.reader, binary.LittleEndian, &o.ui8); err != nil {
		return err
	}
	header.ChannelCount = o.ui8

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

// parseOGGVorbisCommentHeader parses the Vorbis Comment tags in an Ogg Vorbis file
func (o *oggVorbisParser) parseOGGVorbisCommentHeader() error {
	// Read OGGVorbis page header, specifying false to check the capture pattern
	if _, err := o.parseOGGVorbisPageHeader(false); err != nil {
		return err
	}

	// Parse common header
	headerType, err := o.parseOGGVorbisCommonHeader()
	if err != nil {
		return err
	}

	// Verify header type (3: Vorbis Comment)
	if headerType != byte(3) {
		return TagError{
			Err:     errInvalidStream,
			Format:  o.Format(),
			Details: "invalid header type for Vorbis comment header",
		}
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

// parseOGGVorbisDuration reads out the rest of the file to find the last Ogg Vorbis page header, which
// contains information needed to parse the file duration
func (o *oggVorbisParser) parseOGGVorbisDuration() error {
	// Seek as far forward as sanely possible so we don't need to read tons of excess data
	// For now, a value of 4096 bytes before the end appears to work, and should give a bit
	// of wiggle-room without causing us to read the entire file
	if _, err := o.reader.Seek(-4096, 2); err != nil {
		return err
	}

	// Read the rest of the file to find the last page header
	vorbisFile, err := ioutil.ReadAll(o.reader)
	if err != nil {
		return err
	}

	// Find the index of the last OGGVorbis page header
	index := bytes.LastIndex(vorbisFile, oggMagicNumber)
	if index == -1 {
		return TagError{
			Err:     errInvalidStream,
			Format:  o.Format(),
			Details: "could not detect final Ogg page header",
		}
	}

	// Read using the in-memory bytes to grab the last page header information
	o.reader = bytes.NewReader(vorbisFile[index:])
	pageHeader, err := o.parseOGGVorbisPageHeader(false)
	if err != nil {
		return nil
	}

	// Calculate duration using last granule position divided by sample rate
	o.duration = time.Duration(pageHeader.GranulePosition/uint64(o.idHeader.SampleRate)) * time.Second
	return nil
}

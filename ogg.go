package taggolib

import (
	"bytes"
	//"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/eaburns/bit"
)

var (
	// oggMagicNumber is the magic number used to identify a OGG audio stream
	oggMagicNumber = []byte("OggS")
)

// OGGParser represents a OGG audio metadata tag parser
type OGGParser struct {
	idHeader *oggIDHeader
	reader   io.ReadSeeker
	tags     map[string]string
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
	return 16
}

// Bitrate calculates the audio bitrate for this stream
func (o OGGParser) Bitrate() int {
	return 0
}

// Channels returns the number of channels for this stream
func (o OGGParser) Channels() int {
	return 0
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
	return time.Duration(0 * time.Second)
}

// Encoder returns the encoder for this stream
func (o OGGParser) Encoder() string {
	return ""
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
	return 0
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
		reader: reader,
	}

	// Seek forward to ID header
	if _, err := parser.reader.Seek(24, 1); err != nil {
		return nil, err
	}

	// Parse the required ID header
	if err := parser.parseOGGIDHeader(); err != nil {
		return nil, err
	}

	// Return parser
	return parser, nil
}

// parseOGGIDHeader parses the required identification header for an Ogg Vorbis stream
func (o *OGGParser) parseOGGIDHeader() error {
	// Read the first byte to ensure it is an ID header
	first := make([]byte, 1)
	if _, err := o.reader.Read(first); err != nil {
		return err
	}

	// Ensure proper match
	if first[0] != byte(1) {
		return ErrInvalidStream
	}

	// Ensure 'vorbis' identification word is present
	header := make([]byte, 6)
	if _, err := o.reader.Read(header); err != nil {
		return err
	}

	// Ensure proper word is present
	if !bytes.Equal(header, []byte("vorbis")) {
		return ErrInvalidStream
	}

	// Create and use a bit reader to parse the following fields
	//   32 - Vorbis version
	//    8 - Channel count
	//   32 - Sample rate
	//   32 - Maximum bitrate
	//   32 - Nominal bitrate
	//   32 - Minimum bitrate
	//    4 - Blocksize 0
	//    4 - Blocksize 1
	//    1 - Framing flag
	fields, err := bit.NewReader(o.reader).ReadFields(32, 8, 32, 32, 32, 32, 4, 4, 1)
	if err != nil {
		return err
	}

	// Generate ID header
	o.idHeader = &oggIDHeader{
		VorbisVersion: uint32(fields[0]),
		ChannelCount:  uint8(fields[1]),
		MaxBitrate:    uint32(fields[2]),
		NomBitrate:    uint32(fields[3]),
		MinBitrate:    uint32(fields[4]),
		Blocksize0:    uint8(fields[5]),
		Blocksize1:    uint8(fields[6]),
		Framing:       fields[7] == 1,
	}
	fmt.Println(o.idHeader)

	fmt.Println(string(header))
	return nil
}

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

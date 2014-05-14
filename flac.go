package taggolib

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/eaburns/bit"
)

const (
	// flacStreamInfo denotes a STREAMINFO metadata block
	flacStreamInfo = 0
	// flacVorbisComment denotes a VORBISCOMMENT metadata block
	flacVorbisComment = 4
)

var (
	// flacMagicNumber is the magic number used to identify a FLAC audio stream
	flacMagicNumber = []byte("fLaC")
)

// FLACParser represents a FLAC audio metadata tag parser
type FLACParser struct {
	encoder    string
	properties *flacStreamInfoBlock
	reader     io.Reader
	tags       map[string]string
}

// Album returns the Album tag for this stream
func (f FLACParser) Album() string {
	return f.tags[tagAlbum]
}

// AlbumArtist returns the AlbumArtist tag for this stream
func (f FLACParser) AlbumArtist() string {
	return f.tags[tagAlbumArtist]
}

// Artist returns the Artist tag for this stream
func (f FLACParser) Artist() string {
	return f.tags[tagArtist]
}

// BitDepth returns the bits-per-sample of this stream
func (f FLACParser) BitDepth() int {
	return int(f.properties.BitsPerSample)
}

// Channels returns the number of channels for this stream
func (f FLACParser) Channels() int {
	return int(f.properties.ChannelCount)
}

// Checksum returns the checksum for this stream
func (f FLACParser) Checksum() string {
	return f.properties.MD5Checksum
}

// Comment returns the Comment tag for this stream
func (f FLACParser) Comment() string {
	return f.tags[tagComment]
}

// Date returns the Date tag for this stream
func (f FLACParser) Date() string {
	return f.tags[tagDate]
}

// Duration returns the time duration for this stream
func (f FLACParser) Duration() time.Duration {
	return time.Duration(int64(f.properties.SampleCount) / int64(f.SampleRate())) * time.Second
}

// Encoder returns the encoder for this stream
func (f FLACParser) Encoder() string {
	return f.encoder
}

// Format returns the name of the FLAC format
func (f FLACParser) Format() string {
	return "FLAC"
}

// Genre returns the Genre tag for this stream
func (f FLACParser) Genre() string {
	return f.tags[tagGenre]
}

// SampleRate returns the sample rate in Hertz for this stream
func (f FLACParser) SampleRate() int {
	return int(f.properties.SampleRate)
}

// Tag attempts to return the raw, unprocessed tag with the specified name for this stream
func (f FLACParser) Tag(name string) string {
	return f.tags[strings.ToUpper(name)]
}

// Title returns the Title tag for this stream
func (f FLACParser) Title() string {
	return f.tags[tagTitle]
}

// TrackNumber returns the TrackNumber tag for this stream
func (f FLACParser) TrackNumber() int {
	track, err := strconv.Atoi(f.tags[tagTrackNumber])
	if err != nil {
		return 0
	}

	return track
}

// newFLACParser creates a parser for FLAC audio streams
func newFLACParser(reader io.Reader) (*FLACParser, error) {
	// Create FLAC parser
	parser := &FLACParser{
		reader: reader,
	}

	// Begin parsing properties
	if err := parser.parseProperties(); err != nil {
		return nil, err
	}

	// Seek through the file and attempt to parse tags
	if err := parser.parseTags(); err != nil {
		return nil, err
	}

	// Return parser
	return parser, nil
}

// flacMetadataHeader represents the header for a FLAC metadata block
type flacMetadataHeader struct {
	LastBlock   bool
	BlockType   uint8
	BlockLength uint32
}

// flacStreamInfoBlock represents the metadata from a FLAC STREAMINFO block
type flacStreamInfoBlock struct {
	MinBlockSize  uint16
	MaxBlockSize  uint16
	MinFrameSize  uint32
	MaxFrameSize  uint32
	SampleRate    uint16
	ChannelCount  uint8
	BitsPerSample uint16
	SampleCount   uint64
	MD5Checksum   string
}

// parseMetadataHeader retrieves metadata header information from a FLAC stream
func (f *FLACParser) parseMetadataHeader() (*flacMetadataHeader, error) {
	// Create and use a bit reader to parse the following fields:
	//    1 - Last metadata block before audio (boolean)
	//    7 - Metadata block type (should be 0, for streaminfo)
	//   24 - Length of following metadata (in bytes)
	fields, err := bit.NewReader(f.reader).ReadFields(1, 7, 24)
	if err != nil {
		return nil, err
	}

	// Generate metadata header
	return &flacMetadataHeader{
		LastBlock: fields[0] == 1,
		BlockType: uint8(fields[1]),
		BlockLength: uint32(fields[2]),
	}, nil
}

// parseTags retrieves metadata tags from a FLAC VORBISCOMMENT block
func (f *FLACParser) parseTags() error {
	// Continuously parse and seek through blocks until we discover the VORBISCOMMENT block
	for {
		header, err := f.parseMetadataHeader()
		if err != nil {
			return err
		}

		// Check for VORBISCOMMENT block, break so we can begin parsing tags
		if header.BlockType == flacVorbisComment {
			break
		}

		// If last block and no VORBISCOMMENT block found, no tags
		if header.LastBlock {
			return nil
		}

		// If nothing found and not last block, seek forward in stream
		buf := make([]byte, header.BlockLength)
		if _, err := f.reader.Read(buf); err != nil {
			return err
		}
	}

	// Read vendor string length
	var vendorLength uint32
	if err := binary.Read(f.reader, binary.LittleEndian, &vendorLength); err != nil {
		return err
	}

	// Read vendor string
	vendorBuf := make([]byte, vendorLength)
	if _, err := f.reader.Read(vendorBuf); err != nil {
		return err
	}
	f.encoder = string(vendorBuf)

	// Read comment length
	var commentLength uint32
	if err := binary.Read(f.reader, binary.LittleEndian, &commentLength); err != nil {
		return err
	}

	// Begin iterating tags, and building tag map
	tagMap := map[string]string{}
	for i := 0; i < int(commentLength); i++ {
		// Read tag string length
		var tagLength uint32
		if err := binary.Read(f.reader, binary.LittleEndian, &tagLength); err != nil {
			return err
		}

		// Read tag string
		tagBuf := make([]byte, tagLength)
		if _, err := f.reader.Read(tagBuf); err != nil {
			return err
		}

		// Split tag name and data, store in map
		pair := strings.Split(string(tagBuf), "=")
		tagMap[strings.ToUpper(pair[0])] = pair[1]
	}

	// Store tags
	f.tags = tagMap
	return nil
}

// parseProperties retrieves stream properties from a FLAC STREAMINFO block
func (f *FLACParser) parseProperties() error {
	// Read the metadata header for STREAMINFO block
	header, err := f.parseMetadataHeader()
	if err != nil {
		return err
	}

	// Ensure not last field, and that the metadata block type is STREAMINFO
	if header.LastBlock || header.BlockType != flacStreamInfo {
		return ErrInvalidStream
	}

	// Create and use a bit reader to parse the following fields:
	//   16 - Minimum block size (in samples)
	//   16 - Maximum block size (in samples)
	//   24 - Minimum frame size (in bytes)
	//   24 - Maximum frame size (in bytes)
	//   20 - Sample rate
	//    3 - Channel count (+1)
	//    5 - Bits per sample (+1)
	//   36 - Sample count
	fields, err := bit.NewReader(f.reader).ReadFields(16, 16, 24, 24, 20, 3, 5, 36)
	if err != nil {
		return err
	}

	// Read the MD5 checksum of the stream
	checksum := make([]byte, 16)
	if _, err := f.reader.Read(checksum); err != nil {
		return ErrInvalidStream
	}

	// Store properties
	f.properties = &flacStreamInfoBlock{
		MinBlockSize:  uint16(fields[0]),
		MaxBlockSize:  uint16(fields[1]),
		MinFrameSize:  uint32(fields[2]),
		MaxFrameSize:  uint32(fields[3]),
		SampleRate:    uint16(fields[4]),
		ChannelCount:  uint8(fields[5]) + 1,
		BitsPerSample: uint16(fields[6]) + 1,
		SampleCount:   uint64(fields[7]),
		MD5Checksum:   fmt.Sprintf("%x", checksum),
	}

	return nil
}

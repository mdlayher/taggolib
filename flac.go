package taggolib

import (
	"fmt"
	"io"

	"github.com/eaburns/bit"
)

const (
	// flacStreamInfo denotes a STREAMINFO metadata block
	flacStreamInfo = iota
)

var (
	// flacMagicNumber is the magic number used to identify a FLAC audio stream
	flacMagicNumber = []byte("fLaC")
)

// FLACParser represents a FLAC audio metadata tag parser
type FLACParser struct {
	reader     io.Reader
	properties *flacStreamInfoBlock
}

// Format returns the name of the FLAC format
func (f FLACParser) Format() string {
	return "FLAC"
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

	// Return parser
	return parser, nil
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

// parseProperties retrieves stream properties from a FLAC STREAMINFO block
func (f *FLACParser) parseProperties() error {
	// Create and use a bit reader to parse the following fields:
	//    1 - Last metadata block before audio (boolean)
	//    7 - Metadata block type (should be 0, for streaminfo)
	//   24 - Length of following metadata (in bytes)
	//   16 - Minimum block size (in samples)
	//   16 - Maximum block size (in samples)
	//   24 - Minimum frame size (in bytes)
	//   24 - Maximum frame size (in bytes)
	//   20 - Sample rate
	//    3 - Channel count (+1)
	//    5 - Bits per sample (+1)
	//   36 - Sample count
	fields, err := bit.NewReader(f.reader).ReadFields(1, 7, 24, 16, 16, 24, 24, 20, 3, 5, 36)
	if err != nil {
		return err
	}

	// Ensure not last field, and that the metadata block type is STREAMINFO
	if fields[0] == 1 || fields[1] != flacStreamInfo {
		return ErrInvalidStream
	}

	// Read the MD5 checksum of the stream
	checksum := make([]byte, 16)
	if _, err := f.reader.Read(checksum); err != nil {
		return ErrInvalidStream
	}

	// Store properties
	f.properties = &flacStreamInfoBlock{
		MinBlockSize:  uint16(fields[3]),
		MaxBlockSize:  uint16(fields[4]),
		MinFrameSize:  uint32(fields[5]),
		MaxFrameSize:  uint32(fields[6]),
		SampleRate:    uint16(fields[7]),
		ChannelCount:  uint8(fields[8]) + 1,
		BitsPerSample: uint16(fields[9]) + 1,
		SampleCount:   uint64(fields[10]),
		MD5Checksum:   fmt.Sprintf("%x", checksum),
	}

	return nil
}

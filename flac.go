package taggolib

import (
	"bufio"
	"bytes"
	"errors"
)

var (
	// flacMagicNumber is the magic number used to identify a FLAC audio stream
	flacMagicNumber = []byte("fLaC")

	// ErrFLACMagicNumber is returned when an invalid FLAC stream is opened by the FLACParser
	ErrFLACMagicNumber = errors.New("taggolib: invalid FLAC magic number")
)

// FLACParser represents a FLAC audio metadata tag parser
type FLACParser struct{
	buffer *bufio.Reader
}

// Format returns the name of the FLAC format
func (f FLACParser) Format() string {
	return "FLAC"
}

// newFLACParser creates a parser for FLAC audio streams
func newFLACParser(reader *bufio.Reader) (*FLACParser, error) {
	// Read the first 4 bytes to check for FLAC magic number
	magic := make([]byte, 4)
	if _, err := reader.Read(magic); err != nil {
		return nil, err
	}

	// If the proper magic number is not found, this reader does not contain a valid FLAC stream
	if !bytes.Equal(magic, flacMagicNumber) {
		return nil, ErrFLACMagicNumber
	}

	// Return FLAC parser
	return &FLACParser{
		buffer: reader,
	}, nil
}

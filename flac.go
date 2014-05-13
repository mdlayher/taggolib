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
type FLACParser struct{}

// Format returns the name of the FLAC format
func (f FLACParser) Format() string {
	return "FLAC"
}

// NewFLACParser creates a parser for FLAC audio streams
func NewFLACParser(reader *bufio.Reader) (*FLACParser, error) {
	// Peek at the first 4 bytes to check for FLAC magic number
	magic, err := reader.Peek(4)
	if err != nil {
		return nil, err
	}

	// If the proper magic number is not found, this reader does not contain a valid FLAC stream
	if !bytes.Equal(magic, flacMagicNumber) {
		return nil, ErrFLACMagicNumber
	}

	// Return FLAC parser
	return &FLACParser{}, nil
}

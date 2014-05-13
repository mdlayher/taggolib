package taggolib

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

var (
	// ErrUnknownFormat is returned when taggolib cannot recognize the input stream format
	ErrUnknownFormat = errors.New("taggolib: unknown format")
)

// Parser represents an audio metadata tag parser
type Parser interface {
	Format() string
}

// New creates a new Parser depending on the magic number detected in the input reader
func New(reader io.Reader) (Parser, error) {
	// Wrap the raw reader in a buffered one
	bufReader := bufio.NewReader(reader)

	// Peek at the first 4 bytes to check for an audio format magic number
	magic, err := bufReader.Peek(4)
	if err != nil {
		return nil, err
	}

	// Check for FLAC magic number
	if bytes.Equal(magic, flacMagicNumber) {
		return NewFLACParser(bufReader)
	}

	// Check for MP3 magic number
	if bytes.Equal(magic[0:3], mp3MagicNumber) {
		return NewMP3Parser(bufReader)
	}

	// Unrecognized magic number
	return nil, ErrUnknownFormat
}

package taggolib

import (
	"bufio"
	"bytes"
	"errors"
)

var (
	// mp3MagicNumber is the magic number used to identify a MP3 audio stream with ID3V2 tags
	mp3MagicNumber = []byte("ID3")

	// ErrMP3MagicNumber is returned when an invalid MP3 stream is opened by the MP3Parser
	ErrMP3MagicNumber = errors.New("taggolib: invalid MP3 magic number")
)

// MP3Parser represents a MP3 audio metadata tag parser
type MP3Parser struct{
	buffer *bufio.Reader
}

// Format returns the name of the MP3 format
func (m MP3Parser) Format() string {
	return "MP3"
}

// newMP3Parser creates a parser for MP3 audio streams
func newMP3Parser(reader *bufio.Reader) (*MP3Parser, error) {
	// Read the first 3 bytes to check for MP3 magic number
	magic := make([]byte, 3)
	if _, err := reader.Read(magic); err != nil {
		return nil, err
	}

	// If the proper magic number is not found, this reader does not contain a valid MP3 stream
	if !bytes.Equal(magic, mp3MagicNumber) {
		return nil, ErrMP3MagicNumber
	}

	// Return MP3 parser with readed embedded
	return &MP3Parser{
		buffer: reader,
	}, nil
}

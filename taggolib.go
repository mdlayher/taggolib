package taggolib

import (
	"bytes"
	"errors"
	"io"
	"time"
)

const (
	// These constants represent the built-in tags
	tagAlbum = "ALBUM"
	tagAlbumArtist = "ALBUMARTIST"
	tagArtist = "ARTIST"
	tagComment = "COMMENT"
	tagDate = "DATE"
	tagGenre = "GENRE"
	tagTitle = "TITLE"
	tagTrackNumber = "TRACKNUMBER"
)

var (
	// ErrInvalidStream is returned when taggolib encounters a broken input stream
	ErrInvalidStream = errors.New("taggolib: invalid input stream")
	// ErrUnknownFormat is returned when taggolib cannot recognize the input stream format
	ErrUnknownFormat = errors.New("taggolib: unknown format")
)

// Parser represents an audio metadata tag parser
type Parser interface {
	Album() string
	AlbumArtist() string
	Artist() string
	BitDepth() int
	Channels() int
	Checksum() string
	Comment() string
	Date() string
	Duration() time.Duration
	Format() string
	Genre() string
	SampleRate() int
	Tag(name string) string
	Title() string
	TrackNumber() int
}

// New creates a new Parser depending on the magic number detected in the input reader
func New(reader io.Reader) (Parser, error) {
	// Read first byte to begin checking magic number
	first := make([]byte, 1)
	if _, err := reader.Read(first); err != nil {
		return nil, err
	}

	// Check for FLAC magic number
	if bytes.Equal(first, []byte("f")) {
		// Read next 3 bytes for magic number
		magic := make([]byte, 3)
		if _, err := reader.Read(magic); err != nil {
			return nil, err
		}

		// Verify FLAC magic number
		if bytes.Equal(append(first, magic...), flacMagicNumber) {
			return newFLACParser(reader)
		}
	}

	// Unrecognized magic number
	return nil, ErrUnknownFormat
}

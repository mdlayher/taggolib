package taggolib

import (
	"bytes"
	"errors"
	"io"
	"time"
)

const (
	// These constants represent the built-in tags
	tagAlbum       = "ALBUM"
	tagAlbumArtist = "ALBUMARTIST"
	tagArtist      = "ARTIST"
	tagComment     = "COMMENT"
	tagDate        = "DATE"
	tagDiscNumber  = "DISCNUMBER"
	tagGenre       = "GENRE"
	tagTitle       = "TITLE"
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
	Bitrate() int
	Channels() int
	Comment() string
	Date() string
	DiscNumber() int
	Duration() time.Duration
	Encoder() string
	Format() string
	Genre() string
	SampleRate() int
	Tag(name string) string
	Title() string
	TrackNumber() int
}

// New creates a new Parser depending on the magic number detected in the input reader
func New(reader io.ReadSeeker) (Parser, error) {
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

	// Check for MP3 magic number
	if bytes.Equal(first, []byte("I")) {
		// Read next 2 bytes for magic number
		magic := make([]byte, 2)
		if _, err := reader.Read(magic); err != nil {
			return nil, err
		}

		// Verify MP3 magic number
		if bytes.Equal(append(first, magic...), mp3MagicNumber) {
			return newMP3Parser(reader)
		}
	}

	// Check for OGG magic number
	if bytes.Equal(first, []byte("O")) {
		// Read next 3 bytes for magic number
		magic := make([]byte, 3)
		if _, err := reader.Read(magic); err != nil {
			return nil, err
		}

		// Verify OGG magic number
		if bytes.Equal(append(first, magic...), oggMagicNumber) {
			return newOGGParser(reader)
		}
	}

	// Unrecognized magic number
	return nil, ErrUnknownFormat
}

package taggolib

import (
	"bytes"
	"errors"
	"fmt"
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
	// errInvalidStream is returned when taggolib encounters a broken input stream, but
	// does recognize the input stream format
	errInvalidStream = errors.New("invalid input stream")
	// errUnknownFormat is returned when taggolib cannot recognize the input stream format
	errUnknownFormat = errors.New("unknown format")
	// errUnsupportedVersion is returned when taggolib recognizes an input stream format, but
	// can not currently handle the version specified by the input stream
	errUnsupportedVersion = errors.New("unsupported version")
)

// TagError represents an error which occurs during the metadata parsing process
type TagError struct {
	Err     error
	Format  string
	Details string
}

// Error returns a detailed description of an error during the the metadata parsing process
func (e TagError) Error() string {
	return fmt.Sprintf("%s - %s: %s", e.Err.Error(), e.Format, e.Details)
}

// IsInvalidStream is a convenience method which checks if an error is caused by an invalid stream
// of a known format
func IsInvalidStream(err error) bool {
	// Attempt to type-assert to TagError
	tagErr, ok := err.(TagError)
	if !ok {
		return false
	}

	// Return if error matches errInvalidStream
	return tagErr.Err == errInvalidStream
}

// IsUnknownFormat is a convenience method which checks if an error is caused by an unknown format
func IsUnknownFormat(err error) bool {
	// Attempt to type-assert to TagError
	tagErr, ok := err.(TagError)
	if !ok {
		return false
	}

	// Return if error matches errUnknownFormat
	return tagErr.Err == errUnknownFormat
}

// IsUnsupportedVersion is a convenience method which checks if an error is caused by an unsupported version
// of a known format
func IsUnsupportedVersion(err error) bool {
	// Attempt to type-assert to TagError
	tagErr, ok := err.(TagError)
	if !ok {
		return false
	}

	// Return if error matches errUnsupportedVersion
	return tagErr.Err == errUnsupportedVersion
}

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
	// Check for magic numbers
	magicBuf := make([]byte, 8)

	// Read first byte to begin checking magic number
	if _, err := reader.Read(magicBuf[:1]); err != nil {
		return nil, err
	}

	// Check for FLAC magic number
	if magicBuf[0] == byte('f') {
		// Read next 3 bytes for magic number
		if _, err := reader.Read(magicBuf[1:len(flacMagicNumber)]); err != nil {
			return nil, err
		}

		// Verify FLAC magic number
		if bytes.Equal(magicBuf[:len(flacMagicNumber)], flacMagicNumber) {
			return newFLACParser(reader)
		}
	}

	// Check for MP3 magic number
	if magicBuf[0] == byte('I') {
		// Read next 2 bytes for magic number
		if _, err := reader.Read(magicBuf[1:len(mp3MagicNumber)]); err != nil {
			return nil, err
		}

		// Verify MP3 magic number
		if bytes.Equal(magicBuf[:len(mp3MagicNumber)], mp3MagicNumber) {
			return newMP3Parser(reader)
		}
	}

	// Check for OGG magic number
	if magicBuf[0] == byte('O') {
		// Read next 3 bytes for magic number
		if _, err := reader.Read(magicBuf[1:len(oggMagicNumber)]); err != nil {
			return nil, err
		}

		// Verify OGG magic number
		if bytes.Equal(magicBuf[:len(oggMagicNumber)], oggMagicNumber) {
			return newOGGVorbisParser(reader)
		}
	}

	// Unrecognized magic number
	return nil, TagError{
		Err:     errUnknownFormat,
		Format:  "unknown",
		Details: "unrecognized magic number, cannot parse this stream",
	}
}

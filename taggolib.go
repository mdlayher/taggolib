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
	tagPublisher   = "PUBLISHER"
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

// TagError represents an error which occurs during the metadata parsing process.  It is used internally to
// note several types of errors, and may be used to retrieve detailed information regarding an error.
type TagError struct {
	Err     error
	Format  string
	Details string
}

// Error returns a detailed description of an error during the the metadata parsing process, including the
// internal taggolib error, the detected stream format, and a short description of exactly why the error occurred.
func (e TagError) Error() string {
	return fmt.Sprintf("%s - %s: %s", e.Err.Error(), e.Format, e.Details)
}

// IsInvalidStream is a convenience method which checks if an error is caused by an invalid stream
// of a known format.  This may happen if the input stream is corrupt, or if the input stream contains flags which
// should not be present in a valid input stream.
func IsInvalidStream(err error) bool {
	// Attempt to type-assert to TagError
	tagErr, ok := err.(TagError)
	if !ok {
		return false
	}

	// Return if error matches errInvalidStream
	return tagErr.Err == errInvalidStream
}

// IsUnknownFormat is a convenience method which checks if an error is caused by an unknown format.  This may happen
// if the input stream contains a magic number which taggolib cannot handle, such as an unsupported audio format,
// or any kind of file which is not an audio file.
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
// of a known format.  This may happen if the input stream is recognized by taggolib, but taggolib does not support
// parsing a certain version of the metadata, such as ID3v1.
func IsUnsupportedVersion(err error) bool {
	// Attempt to type-assert to TagError
	tagErr, ok := err.(TagError)
	if !ok {
		return false
	}

	// Return if error matches errUnsupportedVersion
	return tagErr.Err == errUnsupportedVersion
}

// Parser represents an audio metadata tag parser.  It is the interface which all other parsers implement, and it
// contains all the standard methods which must be present in an audio parser.
type Parser interface {
	// Methods which access the data stored in a typical audio metadata tag
	Album() string
	AlbumArtist() string
	Artist() string
	Comment() string
	Date() string
	DiscNumber() int
	Genre() string
	Publisher() string
	Title() string
	TrackNumber() int

	// Tag is a special method which will attempt to retrieve an audio metadata
	// tag with the input name. Tag will attempt to return a metadata tag's raw
	// contents, or will return an empty string on failure.
	// Using Tag, the following two calls are functionally equivalent:
	//   - parser.Artist()
	//   - parser.Tag("ARTIST")
	Tag(name string) string

	// Methods which access properties of an audio file, which are
	// typically calculated at runtime
	BitDepth() int
	Bitrate() int
	Channels() int
	Duration() time.Duration
	Encoder() string
	Format() string
	SampleRate() int
}

// New creates a new audio metadata parser, depending on the magic number detected in the input reader.  If New
// recognizes the magic number, it will delegate parsing to the appropriate parser.  If it does not recognize the
// input format, it will return errUnknownFormat, which can be checked using IsUnknownFormat.
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

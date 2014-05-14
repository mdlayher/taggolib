package taggolib

import (
	"io"
	"time"
)

var (
	// mp3MagicNumber is the magic number used to identify a MP3 audio stream
	mp3MagicNumber = []byte("ID3")
)

// MP3Parser represents a MP3 audio metadata tag parser
type MP3Parser struct {
	reader io.ReadSeeker
}

// Album returns the Album tag for this stream
func (m MP3Parser) Album() string {
	return ""
}

// AlbumArtist returns the AlbumArtist tag for this stream
func (m MP3Parser) AlbumArtist() string {
	return ""
}

// Artist returns the Artist tag for this stream
func (m MP3Parser) Artist() string {
	return ""
}

// BitDepth returns the bits-per-sample of this stream
func (m MP3Parser) BitDepth() int {
	return 0
}

// Bitrate calculates the audio bitrate for this stream
func (m MP3Parser) Bitrate() int {
	return 0
}

// Channels returns the number of channels for this stream
func (m MP3Parser) Channels() int {
	return 0
}

// Checksum returns the checksum for this stream
func (m MP3Parser) Checksum() string {
	return ""
}

// Comment returns the Comment tag for this stream
func (m MP3Parser) Comment() string {
	return ""
}

// Date returns the Date tag for this stream
func (m MP3Parser) Date() string {
	return ""
}

// Duration returns the time duration for this stream
func (m MP3Parser) Duration() time.Duration {
	return time.Duration(1 * time.Second)
}

// Encoder returns the encoder for this stream
func (m MP3Parser) Encoder() string {
	return ""
}

// Format returns the name of the MP3 format
func (m MP3Parser) Format() string {
	return "MP3"
}

// Genre returns the Genre tag for this stream
func (m MP3Parser) Genre() string {
	return ""
}

// SampleRate returns the sample rate in Hertz for this stream
func (m MP3Parser) SampleRate() int {
	return 0
}

// Tag attempts to return the raw, unprocessed tag with the specified name for this stream
func (m MP3Parser) Tag(name string) string {
	return ""
}

// Title returns the Title tag for this stream
func (m MP3Parser) Title() string {
	return ""
}

// TrackNumber returns the TrackNumber tag for this stream
func (m MP3Parser) TrackNumber() int {
	return 0
}

// newMP3Parser creates a parser for MP3 audio streams
func newMP3Parser(reader io.ReadSeeker) (*MP3Parser, error) {
	// Create MP3 parser
	parser := &MP3Parser{
		reader: reader,
	}

	// Return parser
	return parser, nil
}

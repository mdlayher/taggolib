package taggolib

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/eaburns/bit"
)

const (
	// flacStreamInfo denotes a STREAMINFO metadata block
	flacStreamInfo = 0
	// flacVorbisComment denotes a VORBISCOMMENT metadata block
	flacVorbisComment = 4
)

var (
	// flacMagicNumber is the magic number used to identify a FLAC audio stream
	flacMagicNumber = []byte("fLaC")
)

// flacParser represents a FLAC audio metadata tag parser
type flacParser struct {
	encoder    string
	endPos     int64
	properties *flacStreamInfoBlock
	reader     io.ReadSeeker
	tags       map[string]string

	// Shared buffer stored as field to prevent unneeded allocations
	buffer []byte
}

// Album returns the Album tag for this stream
func (f flacParser) Album() string {
	return f.tags[tagAlbum]
}

// AlbumArtist returns the AlbumArtist tag for this stream
func (f flacParser) AlbumArtist() string {
	return f.tags[tagAlbumArtist]
}

// Artist returns the Artist tag for this stream
func (f flacParser) Artist() string {
	return f.tags[tagArtist]
}

// BitDepth returns the bits-per-sample of this stream
func (f flacParser) BitDepth() int {
	return int(f.properties.BitsPerSample)
}

// Bitrate calculates the audio bitrate for this stream
func (f flacParser) Bitrate() int {
	return int(((f.endPos * 8) / int64(f.Duration().Seconds())) / 1024)
}

// Channels returns the number of channels for this stream
func (f flacParser) Channels() int {
	return int(f.properties.ChannelCount)
}

// Checksum returns the checksum for this stream
func (f flacParser) Checksum() string {
	return f.properties.MD5Checksum
}

// Comment returns the Comment tag for this stream
func (f flacParser) Comment() string {
	return f.tags[tagComment]
}

// Date returns the Date tag for this stream
func (f flacParser) Date() string {
	return f.tags[tagDate]
}

// DiscNumber returns the DiscNumber tag for this stream
func (f flacParser) DiscNumber() int {
	disc, err := strconv.Atoi(f.tags[tagDiscNumber])
	if err != nil {
		return 0
	}

	return disc
}

// Duration returns the time duration for this stream
func (f flacParser) Duration() time.Duration {
	return time.Duration(int64(f.properties.SampleCount)/int64(f.SampleRate())) * time.Second
}

// Encoder returns the encoder for this stream
func (f flacParser) Encoder() string {
	return f.encoder
}

// Format returns the name of the FLAC format
func (f flacParser) Format() string {
	return "FLAC"
}

// Genre returns the Genre tag for this stream
func (f flacParser) Genre() string {
	return f.tags[tagGenre]
}

// SampleRate returns the sample rate in Hertz for this stream
func (f flacParser) SampleRate() int {
	return int(f.properties.SampleRate)
}

// Tag attempts to return the raw, unprocessed tag with the specified name for this stream
func (f flacParser) Tag(name string) string {
	return f.tags[strings.ToUpper(name)]
}

// Title returns the Title tag for this stream
func (f flacParser) Title() string {
	return f.tags[tagTitle]
}

// TrackNumber returns the TrackNumber tag for this stream
func (f flacParser) TrackNumber() int {
	track, err := strconv.Atoi(f.tags[tagTrackNumber])
	if err != nil {
		return 0
	}

	return track
}

// newFLACParser creates a parser for FLAC audio streams
func newFLACParser(reader io.ReadSeeker) (*flacParser, error) {
	// Create FLAC parser
	parser := &flacParser{
		buffer: make([]byte, 128),
		reader: reader,
	}

	// Begin parsing properties
	if err := parser.parseProperties(); err != nil {
		return nil, err
	}

	// Seek through the file and attempt to parse tags
	if err := parser.parseTags(); err != nil {
		return nil, err
	}

	// Seek to end of file to grab the final position, used to calculate bitrate
	n, err := parser.reader.Seek(0, 2)
	if err != nil {
		return nil, err
	}
	parser.endPos = n

	// Return parser
	return parser, nil
}

// flacMetadataHeader represents the header for a FLAC metadata block
type flacMetadataHeader struct {
	LastBlock   bool
	BlockType   uint8
	BlockLength uint32
}

// flacStreamInfoBlock represents the metadata from a FLAC STREAMINFO block
type flacStreamInfoBlock struct {
	SampleRate    uint16
	ChannelCount  uint8
	BitsPerSample uint16
	SampleCount   uint64
	MD5Checksum   string
}

// parseMetadataHeader retrieves metadata header information from a FLAC stream
func (f *flacParser) parseMetadataHeader() (*flacMetadataHeader, error) {
	// Create and use a bit reader to parse the following fields:
	//    1 - Last metadata block before audio (boolean)
	//    7 - Metadata block type (should be 0, for streaminfo)
	//   24 - Length of following metadata (in bytes)
	fields, err := bit.NewReader(f.reader).ReadFields(1, 7, 24)
	if err != nil {
		return nil, err
	}

	// Generate metadata header
	return &flacMetadataHeader{
		LastBlock:   fields[0] == 1,
		BlockType:   uint8(fields[1]),
		BlockLength: uint32(fields[2]),
	}, nil
}

// parseTags retrieves metadata tags from a FLAC VORBISCOMMENT block
func (f *flacParser) parseTags() error {
	// Continuously parse and seek through blocks until we discover the VORBISCOMMENT block
	for {
		header, err := f.parseMetadataHeader()
		if err != nil {
			return err
		}

		// Check for VORBISCOMMENT block, break so we can begin parsing tags
		if header.BlockType == flacVorbisComment {
			break
		}

		// If last block and no VORBISCOMMENT block found, no tags
		if header.LastBlock {
			return nil
		}

		// If nothing found and not last block, seek forward in stream
		if _, err := f.reader.Seek(int64(header.BlockLength), 1); err != nil {
			return err
		}
	}

	// Parse length fields
	var length uint32

	// Read vendor string length
	if err := binary.Read(f.reader, binary.LittleEndian, &length); err != nil {
		return err
	}

	// Read vendor string
	if _, err := f.reader.Read(f.buffer[:length]); err != nil {
		return err
	}
	f.encoder = string(f.buffer[:length])

	// Read comment length (new allocation so we can use it as loop counter)
	var commentLength uint32
	if err := binary.Read(f.reader, binary.LittleEndian, &commentLength); err != nil {
		return err
	}

	// Begin iterating tags, and building tag map
	tagMap := map[string]string{}
	for i := 0; i < int(commentLength); i++ {
		// Read tag string length
		if err := binary.Read(f.reader, binary.LittleEndian, &length); err != nil {
			return err
		}

		// Read tag string
		n, err := f.reader.Read(f.buffer[:length])
		if err != nil {
			return err
		}

		// Split tag name and data, store in map
		pair := strings.Split(string(f.buffer[:n]), "=")
		tagMap[strings.ToUpper(pair[0])] = pair[1]
	}

	// Store tags
	f.tags = tagMap
	return nil
}

// parseProperties retrieves stream properties from a FLAC STREAMINFO block
func (f *flacParser) parseProperties() error {
	// Read the metadata header for STREAMINFO block
	header, err := f.parseMetadataHeader()
	if err != nil {
		return err
	}

	// Ensure not last field, and that the metadata block type is STREAMINFO
	if header.LastBlock || header.BlockType != flacStreamInfo {
		return ErrInvalidStream
	}

	// Seek forward past frame information, to sample rate
	if _, err := f.reader.Seek(10, 1); err != nil {
		return err
	}

	// Create and use a bit reader to parse the following fields:
	//   20 - Sample rate
	//    3 - Channel count (+1)
	//    5 - Bits per sample (+1)
	//   36 - Sample count
	fields, err := bit.NewReader(f.reader).ReadFields(20, 3, 5, 36)
	if err != nil {
		return err
	}

	// Read the MD5 checksum of the stream
	if _, err := f.reader.Read(f.buffer[:16]); err != nil {
		return ErrInvalidStream
	}

	// Store properties
	f.properties = &flacStreamInfoBlock{
		SampleRate:    uint16(fields[0]),
		ChannelCount:  uint8(fields[1]) + 1,
		BitsPerSample: uint16(fields[2]) + 1,
		SampleCount:   uint64(fields[3]),
		MD5Checksum:   fmt.Sprintf("%x", f.buffer[:16]),
	}

	return nil
}

package taggolib

import (
	"bytes"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
)

var (
	// Read in test files
	flacFile = func() []byte {
		file, err := ioutil.ReadFile("./test/tone16bit.flac")
		if err != nil {
			log.Fatalf("Could not open test FLAC: %v", err)
		}

		return file
	}()
	mp3ID3v23File = func() []byte {
		file, err := ioutil.ReadFile("./test/tone16bit_id3v2.3.mp3")
		if err != nil {
			log.Fatalf("Could not open test MP3 + ID3v2.3: %v", err)
		}

		return file
	}()
	mp3ID3v24File = func() []byte {
		file, err := ioutil.ReadFile("./test/tone16bit_id3v2.4.mp3")
		if err != nil {
			log.Fatalf("Could not open test MP3 + ID3v2.4: %v", err)
		}

		return file
	}()
	oggVorbisFile = func() []byte {
		file, err := ioutil.ReadFile("./test/tone16bit.ogg")
		if err != nil {
			log.Fatalf("Could not open test Ogg Vorbis: %v", err)
		}

		return file
	}()
)

// TestNew verifies that New creates the proper parser for an example input stream
func TestNew(t *testing.T) {
	// Table of tests
	var tests = []struct {
		stream     []byte
		parser     Parser
		err        error
		encoder    string
		tags       []string
		properties []int
	}{
		// Check for FLAC file, with hardcoded expected tags and properties
		{flacFile, &flacParser{}, nil, "reference libFLAC 1.1.4 20070213", []string{"Artist", "Album", "Title"}, []int{5, 202, 16, 44100}},

		// Check for MP3 + ID3v2.3 file, with hardcoded expected tags and properties
		{mp3ID3v23File, &mp3Parser{}, nil, "Lavf53.21.1", []string{"Artist", "Album", "Title"}, []int{5, 32, 16, 44100}},

		// Check for MP3 + ID3v2.4 file, with hardcoded expected tags and properties
		{mp3ID3v24File, &mp3Parser{}, nil, "MP3FS", []string{"Artist", "Album", "Title"}, []int{5, 320, 16, 44100}},

		// Check for Ogg Vorbis file, with hardcoded expected tags and properties
		{oggVorbisFile, &oggVorbisParser{}, nil, "Lavf53.21.1", []string{"Artist", "Album", "Title"}, []int{5, 192, 16, 44100}},

		// Check for an unknown format
		{[]byte("nonsense"), nil, errUnknownFormat, "", nil, nil},
	}

	// Iterate all tests
	for _, test := range tests {
		// Generate a io.ReadSeeker
		reader := bytes.NewReader(test.stream)

		// Attempt to create a parser for the reader
		parser, err := New(reader)
		if err != nil {
			// If an error occurred, check if it was expected
			if test.err == errUnknownFormat && !IsUnknownFormat(err) {
				t.Fatalf("unexpected error: %v", err)
			}
		}

		// Verify that the proper parser type was created
		if reflect.TypeOf(parser) != reflect.TypeOf(test.parser) {
			t.Fatalf("unexpected parser type: %v", reflect.TypeOf(parser))
		}

		// Discard nil parser
		if parser == nil {
			continue
		}

		// Check for valid encoder
		if parser.Encoder() != test.encoder {
			t.Fatalf("mismatched Encoder: %v != %v", parser.Encoder(), test.encoder)
		}

		// Check for valid tags
		if test.tags != nil {
			// Artist
			if parser.Artist() != test.tags[0] {
				t.Fatalf("mismatched tag Artist: %v != %v", parser.Artist(), test.tags[0])
			}

			// Album
			if parser.Album() != test.tags[1] {
				t.Fatalf("mismatched tag Album: %v != %v", parser.Album(), test.tags[1])
			}

			// Title
			if parser.Title() != test.tags[2] {
				t.Fatalf("mismatched tag Title: %v != %v", parser.Title(), test.tags[2])
			}
		}

		// Check for valid properties
		if test.properties != nil {
			// Duration
			if int(parser.Duration().Seconds()) != test.properties[0] {
				t.Fatalf("mismatched property Duration: %v != %v", parser.Duration().Seconds(), test.properties[0])
			}

			// Bitrate
			if parser.Bitrate() != test.properties[1] {
				t.Fatalf("mismatched property Bitrate: %v != %v", parser.Bitrate(), test.properties[1])
			}

			// BitDepth
			if parser.BitDepth() != test.properties[2] {
				t.Fatalf("mismatched property BitDepth: %v != %v", parser.BitDepth(), test.properties[2])
			}

			// SampleRate
			if parser.SampleRate() != test.properties[3] {
				t.Fatalf("mismatched property SampleRate: %v != %v", parser.SampleRate(), test.properties[3])
			}
		}
	}
}

// BenchmarkNewFLAC checks the performance of the New() function with a FLAC file
func BenchmarkNewFLAC(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(bytes.NewReader(flacFile))
	}
}

// BenchmarkNewMP3ID3v23 checks the performance of the New() function with a MP3 + ID3v2.3 file
func BenchmarkNewMP3ID3v23(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(bytes.NewReader(mp3ID3v23File))
	}
}

// BenchmarkNewMP3ID3v24 checks the performance of the New() function with a MP3 + ID3v2.4 file
func BenchmarkNewMP3ID3v24(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(bytes.NewReader(mp3ID3v24File))
	}
}

// BenchmarkNewOGGVorbis checks the performance of the New() function with a Ogg Vorbis file
func BenchmarkNewOGGVorbis(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(bytes.NewReader(oggVorbisFile))
	}
}

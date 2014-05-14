package taggolib

import (
	"bytes"
	"reflect"
	"testing"
)

// TestNew verifies that New creates the proper parser for an example input stream
func TestNew(t *testing.T) {
	var tests = []struct {
		stream []byte
		parser Parser
		err    error
	}{
		// Check for FLAC file
		// TODO: add or embed a small FLAC file for testing
		//{flacMagicNumber, &FLACParser{}, nil},

		// Check for an unknown format
		{[]byte("nonsense"), nil, ErrUnknownFormat},
	}

	// Iterate all tests
	for _, test := range tests {
		// Generate a io.Reader
		reader := bytes.NewReader(test.stream)

		// Attempt to create a parser for the reader
		parser, err := New(reader)
		if err != nil {
			// If an error occurred, check if it was expected
			if err != test.err {
				t.Fatalf("unexpected error: %v", err)
			}
		}

		// Verify that the proper parser type was created
		if reflect.TypeOf(parser) != reflect.TypeOf(test.parser) {
			t.Fatalf("unexpected parser type: %v", reflect.TypeOf(parser))
		}
	}
}

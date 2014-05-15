package taggolib

import (
	"bytes"
	"reflect"
	"testing"
)

// TestMP3 verifies that all MP3Parser methods work properly
func TestMP3(t *testing.T) {
	// Generate a MP3Parser
	mp3, err := New(bytes.NewReader(mp3File))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that we actually got a MP3 mp3
	if reflect.TypeOf(mp3) != reflect.TypeOf(&MP3Parser{}) {
		t.Fatalf("unexpected mp3 type: %v", reflect.TypeOf(mp3))
	}

	// Verify all exported methods work properly

	// Album
	if mp3.Album() != "Album" {
		t.Fatalf("mismatched tag Album: %v", mp3.Album())
	}

	// AlbumArtist
	if mp3.AlbumArtist() != "AlbumArtist" {
		t.Fatalf("mismatched tag AlbumArtist: %v", mp3.AlbumArtist())
	}

	// Artist
	if mp3.Artist() != "Artist" {
		t.Fatalf("mismatched tag Artist: %v", mp3.Artist())
	}

	// BitDepth
	if mp3.BitDepth() != 16 {
		t.Fatalf("mismatched property BitDepth: %v", mp3.BitDepth())
	}

	// Bitrate
	if mp3.Bitrate() != 320 {
		t.Fatalf("mismatched property Bitrate: %v", mp3.Bitrate())
	}

	// Channels
	if mp3.Channels() != 2 {
		t.Fatalf("mismatched property Channels: %v", mp3.Channels())
	}

	// Comment
	if mp3.Comment() != "" {
		t.Fatalf("mismatched tag Comment: %v", mp3.Comment())
	}

	// Date
	if mp3.Date() != "2014-01-01" {
		t.Fatalf("mismatched tag Date: %v", mp3.Date())
	}

	// DiscNumber
	if mp3.DiscNumber() != 1 {
		t.Fatalf("mismatched tag DiscNumber: %v", mp3.DiscNumber())
	}

	// Duration
	if int(mp3.Duration().Seconds()) != 5 {
		t.Fatalf("mismatched property Duration: %v", mp3.Duration().Seconds())
	}

	// Encoder
	if mp3.Encoder() != "MP3FS" {
		t.Fatalf("mismatched property Encoder: %v", mp3.Encoder())
	}

	// Format
	if mp3.Format() != "MP3" {
		t.Fatalf("mismatched property Format: %v", mp3.Format())
	}

	// Genre
	if mp3.Genre() != "Genre" {
		t.Fatalf("mismatched tag Genre: %v", mp3.Genre())
	}

	// SampleRate
	if mp3.SampleRate() != 44100 {
		t.Fatalf("mismatched property SampleRate: %v", mp3.SampleRate())
	}

	// Title
	if mp3.Title() != "Title" {
		t.Fatalf("mismatched tag Title: %v", mp3.Title())
	}

	// TrackNumber
	if mp3.TrackNumber() != 1 {
		t.Fatalf("mismatched tag TrackNumber: %v", mp3.TrackNumber())
	}

	// Check a few raw tags

	if mp3.Tag("ARTIST") != "Artist" {
		t.Fatalf("unexpected raw tag ARTIST: %v", mp3.Tag("ARTIST"))
	}

	if mp3.Tag("ALBUM") != "Album" {
		t.Fatalf("unexpected raw tag ALBUM: %v", mp3.Tag("ALBUM"))
	}

	if mp3.Tag("TITLE") != "Title" {
		t.Fatalf("unexpected raw tag TITLE: %v", mp3.Tag("TITLE"))
	}

	// Check a non-existant tag
	if mp3.Tag("NOTEXISTS") != "" {
		t.Fatalf("unexpected raw tag NOTEXISTS: %v", mp3.Tag("NOTEXISTS"))
	}
}

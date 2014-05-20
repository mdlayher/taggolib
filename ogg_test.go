package taggolib

import (
	"bytes"
	"reflect"
	"testing"
)

// TestOGG verifies that all oggParser methods work properly
func TestOGG(t *testing.T) {
	// Generate a oggParser
	ogg, err := New(bytes.NewReader(oggFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that we actually got a OGG ogg
	if reflect.TypeOf(ogg) != reflect.TypeOf(&oggParser{}) {
		t.Fatalf("unexpected ogg type: %v", reflect.TypeOf(ogg))
	}

	// Verify all exported methods work properly

	// Album
	if ogg.Album() != "Album" {
		t.Fatalf("mismatched tag Album: %v", ogg.Album())
	}

	// AlbumArtist
	if ogg.AlbumArtist() != "AlbumArtist" {
		t.Fatalf("mismatched tag AlbumArtist: %v", ogg.AlbumArtist())
	}

	// Artist
	if ogg.Artist() != "Artist" {
		t.Fatalf("mismatched tag Artist: %v", ogg.Artist())
	}

	// BitDepth
	if ogg.BitDepth() != 16 {
		t.Fatalf("mismatched property BitDepth: %v", ogg.BitDepth())
	}

	// Bitrate
	if ogg.Bitrate() != 192 {
		t.Fatalf("mismatched property Bitrate: %v", ogg.Bitrate())
	}

	// Channels
	if ogg.Channels() != 2 {
		t.Fatalf("mismatched property Channels: %v", ogg.Channels())
	}

	// Comment
	if ogg.Comment() != "Comment" {
		t.Fatalf("mismatched tag Comment: %v", ogg.Comment())
	}

	// Date
	if ogg.Date() != "2014-01-01" {
		t.Fatalf("mismatched tag Date: %v", ogg.Date())
	}

	// DiscNumber
	if ogg.DiscNumber() != 1 {
		t.Fatalf("mismatched tag DiscNumber: %v", ogg.DiscNumber())
	}

	// Duration
	if int(ogg.Duration().Seconds()) != 5 {
		t.Fatalf("mismatched property Duration: %v", ogg.Duration().Seconds())
	}

	// Encoder
	if ogg.Encoder() != "Lavf53.21.1" {
		t.Fatalf("mismatched property Encoder: %v", ogg.Encoder())
	}

	// Format
	if ogg.Format() != "OGG" {
		t.Fatalf("mismatched property Format: %v", ogg.Format())
	}

	// Genre
	if ogg.Genre() != "Genre" {
		t.Fatalf("mismatched tag Genre: %v", ogg.Genre())
	}

	// SampleRate
	if ogg.SampleRate() != 44100 {
		t.Fatalf("mismatched property SampleRate: %v", ogg.SampleRate())
	}

	// Title
	if ogg.Title() != "Title" {
		t.Fatalf("mismatched tag Title: %v", ogg.Title())
	}

	// TrackNumber
	if ogg.TrackNumber() != 1 {
		t.Fatalf("mismatched tag TrackNumber: %v", ogg.TrackNumber())
	}

	// Check a few raw tags

	if ogg.Tag("ARTIST") != "Artist" {
		t.Fatalf("unexpected raw tag ARTIST: %v", ogg.Tag("ARTIST"))
	}

	if ogg.Tag("ALBUM") != "Album" {
		t.Fatalf("unexpected raw tag ALBUM: %v", ogg.Tag("ALBUM"))
	}

	if ogg.Tag("TITLE") != "Title" {
		t.Fatalf("unexpected raw tag TITLE: %v", ogg.Tag("TITLE"))
	}

	// Check a non-existant tag
	if ogg.Tag("NOTEXISTS") != "" {
		t.Fatalf("unexpected raw tag NOTEXISTS: %v", ogg.Tag("NOTEXISTS"))
	}
}

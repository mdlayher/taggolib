package taggolib

import (
	"bytes"
	"reflect"
	"testing"
)

// TestFLAC verifies that all FLACParser methods work properly
func TestFLAC(t *testing.T) {
	// Generate a FLACParser
	flac, err := New(bytes.NewReader(flacFile))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that we actually got a FLAC flac
	if reflect.TypeOf(flac) != reflect.TypeOf(&FLACParser{}) {
		t.Fatalf("unexpected flac type: %v", reflect.TypeOf(flac))
	}

	// Very all exported methods work properly

	// Album
	if flac.Album() != "Album" {
		t.Fatalf("mismatched tag Album: %v", flac.Album())
	}

	// AlbumArtist
	if flac.AlbumArtist() != "AlbumArtist" {
		t.Fatalf("mismatched tag AlbumArtist: %v", flac.AlbumArtist())
	}

	// Artist
	if flac.Artist() != "Artist" {
		t.Fatalf("mismatched tag Artist: %v", flac.Artist())
	}

	// BitDepth
	if flac.BitDepth() != 16 {
		t.Fatalf("mismatched property BitDepth: %v", flac.BitDepth())
	}

	// Bitrate
	if flac.Bitrate() != 202 {
		t.Fatalf("mismatched property Bitrate: %v", flac.Bitrate())
	}

	// Channels
	if flac.Channels() != 2 {
		t.Fatalf("mismatched property Channels: %v", flac.Channels())
	}

	// Comment
	if flac.Comment() != "Comment" {
		t.Fatalf("mismatched tag Comment: %v", flac.Comment())
	}

	// Date
	if flac.Date() != "2014-01-01" {
		t.Fatalf("mismatched tag Date: %v", flac.Date())
	}

	// DiscNumber
	if flac.DiscNumber() != 1 {
		t.Fatalf("mismatched tag DiscNumber: %v", flac.DiscNumber())
	}

	// Duration
	if int(flac.Duration().Seconds()) != 5 {
		t.Fatalf("mismatched property Duration: %v", flac.Duration().Seconds())
	}

	// Encoder
	if flac.Encoder() != "reference libFLAC 1.1.4 20070213" {
		t.Fatalf("mismatched property Encoder: %v", flac.Encoder())
	}

	// Format
	if flac.Format() != "FLAC" {
		t.Fatalf("mismatched property Format: %v", flac.Format())
	}

	// Genre
	if flac.Genre() != "Genre" {
		t.Fatalf("mismatched tag Genre: %v", flac.Genre())
	}

	// SampleRate
	if flac.SampleRate() != 44100 {
		t.Fatalf("mismatched property SampleRate: %v", flac.SampleRate())
	}

	// Title
	if flac.Title() != "Title" {
		t.Fatalf("mismatched tag Title: %v", flac.Title())
	}

	// TrackNumber
	if flac.TrackNumber() != 1 {
		t.Fatalf("mismatched tag TrackNumber: %v", flac.TrackNumber())
	}

	// Check a few raw tags

	if flac.Tag("ARTIST") != "Artist" {
		t.Fatalf("unexpected raw tag ARTIST: %v", flac.Tag("ARTIST"))
	}

	if flac.Tag("ALBUM") != "Album" {
		t.Fatalf("unexpected raw tag ALBUM: %v", flac.Tag("ALBUM"))
	}

	if flac.Tag("TITLE") != "Title" {
		t.Fatalf("unexpected raw tag TITLE: %v", flac.Tag("TITLE"))
	}

	// Check a non-existant tag
	if flac.Tag("NOTEXISTS") != "" {
		t.Fatalf("unexpected raw tag NOTEXISTS: %v", flac.Tag("NOTEXISTS"))
	}
}

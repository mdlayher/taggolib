taggolib [![Build Status](https://travis-ci.org/mdlayher/taggolib.svg?branch=master)](https://travis-ci.org/mdlayher/taggolib) [![GoDoc](http://godoc.org/github.com/mdlayher/taggolib?status.png)](http://godoc.org/github.com/mdlayher/taggolib)
========

taggolib is a Go package which provides read-only access to metadata contained in various audio formats.  MIT Licensed.

taggolib is inspired by the [TagLib](http://taglib.github.io/) and [taglib-sharp](https://github.com/mono/taglib-sharp/)
projects.  Its goal is to provide read-only metadata access to a variety of audio formats in Go, without the need
to use a TagLib binding.

Currently, taggolib supports the following formats, with some caveats:

- FLAC
- MP3
  - ID3v2.4 tags only (ID3v2.3 in the works)
- OGG
  - Cannot calculate Duration as of now (returns "zero" duration)

Example
=======

taggolib has a very simple interface, and many tags can be accessed by simply calling an appropriately-named
method with no parameters.  Below is an example script which prints out an informative one-line summary of an
audio file's tags and properties.

```go
package main

import (
	"fmt"
	"os"

	"github.com/mdlayher/taggolib"
)

func main() {
	// Ensure parameter was passed
	if len(os.Args) < 2 {
		fmt.Println("taggo: no filename parameter")
		return
	}

	// Attempt to open parameter file
	path := os.Args[1]
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Load file using taggolib
	audio, err := taggolib.New(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Calculate duration in mm:ss format
	seconds := int(audio.Duration().Seconds())
	minutes := seconds / 60
	seconds = seconds - (minutes * 60)

	// ex: Jimmy Eat World - Bleed American - Sweetness [#1.05] [03:40] [FLAC/1016kbps/16bit/44kHz]
	fmt.Printf("%s - %s - %s [#%d.%02d] [%02d:%02d] [%s/%dkbps/%dbit/%dkHz]\n",
		audio.Artist(), audio.Album(), audio.Title(), audio.DiscNumber(), audio.TrackNumber(),
		minutes, seconds, audio.Format(), audio.Bitrate(), audio.BitDepth(), audio.SampleRate()/1000)
}
```

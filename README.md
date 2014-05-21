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

Example
=======

taggolib has a very simple interface, and many tags can be accessed by simply calling an appropriately-named
method with no parameters.  Below is an example script called `taggo`, which can also be found in this repository
at [taggo/taggo.go](https://github.com/mdlayher/taggolib/blob/master/taggo/taggo.go). `taggo` will perform a recursive
walk on a specified parameter directory, and print out information about any media files it recognizes.

```
$ cd taggo
$ go build
$ ./taggo /home/matt/Music/
```

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mdlayher/taggolib"
)

func main() {
	// Ensure at least one parameter was passed
	if len(os.Args) < 2 {
		fmt.Println("taggo: no file path parameter")
		return
	}

	// Invoke a recursive file walk
	err := filepath.Walk(os.Args[1], func(path string, info os.FileInfo, err error) error {
		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Open file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Load file using taggolib
		audio, err := taggolib.New(file)
		if err != nil {
			// Check fur unknown format or unsupported version, skip these
			if err == taggolib.ErrUnknownFormat || err == taggolib.ErrUnsupportedVersion {
				return nil
			}

			return err
		}

		// Calculate duration in mm:ss format
		seconds := int(audio.Duration().Seconds())
		minutes := seconds / 60
		seconds = seconds - (minutes * 60)

		// Print information about file
		fmt.Printf("%s - %s - %s [#%d.%02d] [%02d:%02d] [%s/%dkbps/%dbit/%dkHz]\n",
			audio.Artist(), audio.Album(), audio.Title(), audio.DiscNumber(), audio.TrackNumber(),
			minutes, seconds, audio.Format(), audio.Bitrate(), audio.BitDepth(), audio.SampleRate()/1000)

		return nil
	})

	// Check for walk error
	if err != nil {
		fmt.Println("taggo:", err)
		return
	}
}
```

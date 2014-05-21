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
- Ogg Vorbis

Example
=======

taggolib has a very simple interface, and many tags can be accessed by simply calling an appropriately-named
method with no parameters. A basic example script can be found at [taggo/taggo.go](https://github.com/mdlayher/taggolib/blob/master/taggo/taggo.go).
`taggo` will perform a recursive walk on a specified parameter directory, and print out information about any
media files it recognizes.

```
$ cd taggo
$ go build
$ ./taggo /home/matt/Music/
```


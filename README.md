taggolib [![Build Status](https://travis-ci.org/mdlayher/taggolib.svg?branch=master)](https://travis-ci.org/mdlayher/taggolib) [![GoDoc](http://godoc.org/github.com/mdlayher/taggolib?status.png)](http://godoc.org/github.com/mdlayher/taggolib)
========

taggolib is a Go package which provides read-only access to metadata contained in various audio formats.  MIT Licensed.

taggolib is inspired by the [TagLib](http://taglib.github.io/) and [taglib-sharp](https://github.com/mono/taglib-sharp/)
projects.  Its goal is to provide read-only metadata access to a variety of audio formats in Go, without the need
to use a TagLib binding.

Currently, taggolib supports the following formats, with some caveats:

- FLAC
- MP3
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
Jimmy Eat World - Bleed American - Bleed American [#1.01] [03:01] [FLAC/1060kbps/16bit/44kHz]
Jimmy Eat World - Bleed American - A Praise Chorus [#1.02] [04:03] [FLAC/1040kbps/16bit/44kHz]
Jimmy Eat World - Bleed American - The Middle [#1.03] [02:45] [FLAC/985kbps/16bit/44kHz]
Jimmy Eat World - Bleed American - Your House [#1.04] [04:46] [FLAC/972kbps/16bit/44kHz]
Jimmy Eat World - Bleed American - Sweetness [#1.05] [03:40] [FLAC/1016kbps/16bit/44kHz]
Jimmy Eat World - Bleed American - Hear You Me [#1.06] [04:44] [FLAC/941kbps/16bit/44kHz]
Jimmy Eat World - Bleed American - If You Don't, Don't [#1.07] [04:33] [FLAC/1014kbps/16bit/44kHz]
Jimmy Eat World - Bleed American - Get it Faster [#1.08] [04:21] [FLAC/888kbps/16bit/44kHz]
Jimmy Eat World - Bleed American - Cautioners [#1.09] [05:21] [FLAC/938kbps/16bit/44kHz]
Jimmy Eat World - Bleed American - The Authority Song [#1.10] [03:37] [FLAC/1033kbps/16bit/44kHz]
Jimmy Eat World - Bleed American - My Sundown [#1.11] [05:47] [FLAC/764kbps/16bit/44kHz]
```

// Command taggo is a simple audio tag parser, which is meant to demonstrate the functionality
// of the taggolib package.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jesseward/taggolib"
)

func main() {
	// Ensure at least one parameter was passed
	if len(os.Args) < 2 {
		fmt.Println("taggo: no file path parameter")
		return
	}

	// Verify path actually exists
	if _, err := os.Stat(os.Args[1]); err != nil {
		fmt.Println("taggo:", err)
		return
	}

	// Invoke a recursive file walk on all parameter directories
	for _, arg := range os.Args[1:] {
		err := filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
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
				// Check for unknown format, which will be skipped
				if taggolib.IsUnknownFormat(err) {
					return nil
				}

				// Check for unsupported version, invalid stream, or EOF, which will be logged and skipped
				if taggolib.IsUnsupportedVersion(err) || taggolib.IsInvalidStream(err) || err == io.EOF {
					fmt.Println("taggo:", err, ":", path)
					return nil
				}

				return err
			}

			// Calculate duration in mm:ss format
			seconds := int(audio.Duration().Seconds())
			minutes := seconds / 60
			seconds = seconds - (minutes * 60)

			// Print information about file
			fmt.Printf("%s - %s - %s [#%d.%02d] [%02d:%02d] [%s/%dkbps/%dbit/%dkHz] [%s]\n",
				audio.Artist(), audio.Album(), audio.Title(), audio.DiscNumber(), audio.TrackNumber(),
				minutes, seconds, audio.Format(), audio.Bitrate(), audio.BitDepth(), audio.SampleRate()/1000, audio.Publisher())

			return nil
		})

		// Check for fatal walk error
		if err != nil {
			fmt.Println("taggo: fatal:", err)
			return
		}
	}
}

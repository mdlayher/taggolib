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

	// Verify path actually exists
	if _, err := os.Stat(os.Args[1]); err != nil {
		fmt.Println("taggo:", err)
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

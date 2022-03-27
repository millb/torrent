package testutil

import (
	"bytes"
	"io"
	"log"
	"os"
)

// Compare two files at specified paths.
func Compare(file1 string, file2 string) bool {
	const chunkSize = 128 * 1024

	// Check file size ...

	f1, err := os.Open(file1)
	if err != nil {
		log.Panic(err)
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		log.Panic(err)
	}
	defer f2.Close()

	for {
		b1 := make([]byte, chunkSize)
		_, err1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, err2 := f2.Read(b2)

		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {
				return true
			} else if err1 == io.EOF || err2 == io.EOF {
				return false
			} else {
				log.Panic(err1, err2)
			}
		}

		if !bytes.Equal(b1, b2) {
			return false
		}
	}
}

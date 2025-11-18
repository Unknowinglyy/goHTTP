package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal("error when opening file")
	}
	defer file.Close()

	buff := make([]byte, 8)
	var curLine strings.Builder
	sep := []byte("\n")

	for {
		num, err := file.Read(buff)
		if err != nil {
			// done reading the file
			break
		}

		// split buff on \n's
		// IMPORTANT: only process bytes actually read via [:num]
		parts := bytes.Split(buff[:num], sep)

		if len(parts) == 1 {
			// no \n's, can simply append to curLine
			curLine.Write(parts[0])
		} else {
			// we have a \n, time to write to stdout and reset curLine
			lastElement := parts[len(parts)-1]
			selectedParts := parts[:len(parts)-1]

			joinedBytes := bytes.Join(selectedParts, nil)
			curLine.Write(joinedBytes)

			fmt.Printf("read: %s\n", curLine.String())
			curLine.Reset()

			// everything after newline should be saved for next iteration
			curLine.Write(lastElement)
		}
	}

	// if any residual data, print it
	if len(curLine.String()) > 0 {
		fmt.Printf("read: %s\n", curLine.String())
	}
}

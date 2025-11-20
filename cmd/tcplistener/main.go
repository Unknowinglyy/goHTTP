package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	// make buffered?
	chanStr := make(chan string)

	go func() {
		defer close(chanStr)

		buff := make([]byte, 8)
		var curLine strings.Builder
		sep := []byte("\n")

		for {
			num, err := f.Read(buff)
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

				chanStr <- curLine.String()
				curLine.Reset()

				// everything after newline should be saved for next iteration
				curLine.Write(lastElement)
			}
		}

		// if any residual data, send it
		if len(curLine.String()) > 0 {
			chanStr <- curLine.String()
		}
	}()
	return chanStr
}

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error when setting up listener: %s", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error when setting up connection: %s", err)
		}
		fmt.Println("Connection established")

		chann := getLinesChannel(conn)

		for elem := range chann {
			fmt.Printf("%s\n", elem)
		}
		// happens after channel closes
		fmt.Println("Connection closed")
	}
}

package main

import (
	"fmt"
	"log"
	"net"

	"goHttp/internal/request"
)

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

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error with RequestFromReader: %v", err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HTTPVersion)

		fmt.Println("Connection closed")
	}
}

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// gives you a UDPAddr struct which is needed in DialUDP
	// setting up which addr:port we are sending to
	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		log.Fatalf("error resolving UDP address: %s", err)
	}

	// (addrs refer to addr:port)
	// leave laddr nil as we don't care what port we are sending from on the local machine
	// arguments -> network, local addr (sending from), remote addr (sending to)
	udpConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("error establishing UDP socket: %s", err)
	}
	defer udpConn.Close()

	buff := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := buff.ReadString('\n')
		if err != nil {
			log.Fatalf("error reading line: %s", err)
		}

		_, err = udpConn.Write([]byte(line))
		if err != nil {
			log.Fatalf("error writing line to UDP connection: %s", err)
		}
	}
}

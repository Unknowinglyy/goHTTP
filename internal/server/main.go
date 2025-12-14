package server

import (
	"fmt"
	"net"
	"sync/atomic"
)

var ErrorClosingOfflineServer = fmt.Errorf("trying to close a server that is already closed")

type Server struct {
	running  *atomic.Bool
	listener net.Listener
}

func Serve(port int) (*Server, error) {
	// It accepts a port and starts handling requests that come in. hardcode the response.

	// hardcode this response:

	// Creates a net.Listener and returns a new Server instance. Starts listening for requests inside a goroutine.
	addr := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	aBool := atomic.Bool{}
	aBool.Store(true)

	server := Server{running: &aBool, listener: listener}

	go server.listen()
	return &server, nil
}

func (s *Server) Close() error {
	if !s.running.Swap(false) {
		return ErrorClosingOfflineServer
	}
	fmt.Println("closing server")
	return s.listener.Close()
}

func (s *Server) listen() {
	// Uses a loop to .Accept new connections as they come in, and handles each one in a new goroutine. I used an atomic.Bool to track whether the server is closed or not so that I can ignore connection errors after the server is closed.
	fmt.Println("starting to listen")

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// shutting down via CTRL+C, so error is expected
			if !s.running.Load() {
				fmt.Println("listener closed, not accepting anymore connections")
				return
			}
			fmt.Println("error accepting connection: ", err)
			break
		}
		// remote addr = addr of client
		// local addr = addr of server
		fmt.Printf("new connection accepted from %s (local address: %s)\n",
			conn.RemoteAddr().String(),
			conn.LocalAddr().String())
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	// Handles a single connection by writing the following response and then closing the connection:
	buf := make([]byte, 1024)
	defer conn.Close()

	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("error reading from connection: %v\n", err)
		return
	}
	fmt.Printf("received %d bytes\n", n)

	respStr := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 12\r\n" +
		"\r\n" +
		"Hello World!"

	res := []byte(respStr)

	fmt.Printf("writing response of %d\n", len(res))
	_, err = conn.Write(res)
	if err != nil {
		fmt.Printf("error writing to connection: %v\n", err)
	}
}

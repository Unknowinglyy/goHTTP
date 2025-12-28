package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"goHttp/internal/headers"
	"goHttp/internal/request"
	"goHttp/internal/response"
)

var ErrorClosingOfflineServer = fmt.Errorf("trying to close a server that is already closed")

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	running  *atomic.Bool
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	status  response.StatusCode
	message string
}

func NewHandlerError(stat response.StatusCode, mess string) *HandlerError {
	return &HandlerError{status: stat, message: mess}
}

func Serve(h Handler, port uint16) (*Server, error) {
	// It accepts a port and starts handling requests that come in.
	// Creates a net.Listener and returns a new Server instance. Starts listening for requests inside a goroutine.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	var aBool atomic.Bool
	aBool.Store(true)

	server := Server{running: &aBool, listener: listener, handler: h}

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
	// uses a loop to .Accept new connections as they come in, and handles each one in a new goroutine.
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
			continue
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
	// Handles a single connection by parsing request from connection,
	// writing a response, and then closing the connection
	defer conn.Close()

	writer := response.NewWriter(conn)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		// write back a minimal response when we can not parse the request
		body := response.StatusBadBody
		heads := response.GetDefaultHeaders(len(body))
		err := heads.Replace("Content-Type", "text/html")
		if err != nil {
			fmt.Printf("error replacing header: %v", err)
			return
		}

		WriteResponse(writer, response.StatusBad, heads, body)
		return
	}

	s.handler(writer, req)
}

func WriteResponse(w *response.Writer, status response.StatusCode, heads headers.Headers, body string) {
	err := w.WriteStatusLine(status)
	if err != nil {
		fmt.Printf("error writing status line: %v\n", err)
		return
	}

	err = w.WriteHeaders(heads)
	if err != nil {
		fmt.Printf("error writing headers: %v\n", err)
		return
	}

	if len(body) != 0 {
		_, err := w.WriteBody([]byte(body))
		if err != nil {
			fmt.Printf("error writing body: %v\n", err)
		}
	}
}

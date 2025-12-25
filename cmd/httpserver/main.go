package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"goHttp/internal/request"
	"goHttp/internal/response"
	"goHttp/internal/server"
)

const port = 42069

func tempHandle(w io.Writer, req *request.Request) *server.HandlerError {
	stat := response.StatusOK
	mess := ""
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		stat = response.StatusBad
		mess = "Your problem is not my problem\n"
	case "/myproblem":
		stat = response.StatusInServErr
		mess = "my bad\n"
	default:
		_, err := w.Write([]byte("all good\n"))
		if err != nil {
			// ignore write errors for now
			return nil
		}
		return nil
	}
	he := server.NewHandlerError(stat, mess)
	return he
}

func main() {
	srv, err := server.Serve(tempHandle, port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Print("Server gracefully stopped\n\n")
}

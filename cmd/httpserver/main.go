package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"goHttp/internal/request"
	"goHttp/internal/response"
	"goHttp/internal/server"
)

const port = 42069

func tempHandle(w *response.Writer, req *request.Request) {
	var status response.StatusCode
	var body string

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		status = response.StatusBad
		body = response.StatusBadBody
	case "/myproblem":
		status = response.StatusInServErr
		body = response.StatusInServErrBody
	default:
		status = response.StatusOK
		body = response.StatusOKBody
	}

	heads := response.GetDefaultHeaders(len(body))
	err := heads.Replace("Content-Type", "text/html")
	if err != nil {
		panic("we should always be able to replace Content-Type")
	}
	// can use this function for conveniency when it comes to testing:
	// server.WriteResponse(w, status, heads, body)
	err = w.WriteStatusLine(status)
	if err != nil {
		fmt.Printf("error writing status line: %v", err)
		return
	}
	err = w.WriteHeaders(heads)
	if err != nil {
		fmt.Printf("error writing headers: %v", err)
		return
	}
	_, err = w.WriteBody([]byte(body))
	if err != nil {
		fmt.Printf("error writing body: %v", err)
		return
	}
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
	<-sigChan // blocking until we get either signal above
	log.Print("Server gracefully stopped\n\n")
}

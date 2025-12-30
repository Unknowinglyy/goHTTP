package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"goHttp/internal/handlers"
	"goHttp/internal/server"
)

const port = 42069

func main() {
	// four handler examples in the internal/handlers package:
	// handler (default handler), chunked encoding handler (proxyHandler),
	// chunked encoding with trailers (proxyHandlerWithTrailers), &
	// default handler that responds with a video file (BinaryDataHandler)

	// srv, err := server.Serve(handlers.Handler, port)
	// srv, err := server.Serve(handlers.ProxyHandlerWithTrailers, port)
	// srv, err := server.Serve(handlers.ProxyHandler, port)
	srv, err := server.Serve(handlers.BinaryDataHandler, port)
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

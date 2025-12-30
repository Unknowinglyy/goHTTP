package handlers

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"

	"goHttp/internal/headers"
	"goHttp/internal/request"
	"goHttp/internal/response"
	"goHttp/internal/server"
)

func Handler(w *response.Writer, req *request.Request) {
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
	err := heads.Update("Content-Type", "text/html")
	if err != nil {
		panic("we should always be able to update Content-Type")
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

func ProxyHandler(w *response.Writer, req *request.Request) {
	httpBinPrefix := "/httpbin/"
	redirect := strings.HasPrefix(req.RequestLine.RequestTarget, httpBinPrefix)

	// make default response if we are not redirecting
	if !redirect {
		body := response.StatusOKBody
		heads := response.GetDefaultHeaders(len(body))
		err := heads.Update("Content-Type", "text/html")
		if err != nil {
			panic("we should always be able to update Content-Type")
		}
		server.WriteResponse(w, response.StatusOK, heads, body)
		return
	}

	// make request to httpbin to get content
	redirTarget := strings.TrimPrefix(req.RequestLine.RequestTarget, httpBinPrefix)
	resp, err := http.Get("https://httpbin.org/" + redirTarget)
	if err != nil {
		fmt.Printf("error getting response from https://httpbin.org/: %v", err)
		return
	}
	defer resp.Body.Close()

	heads := headers.NewHeaders()
	heads.Set("Content-Type", "text/plain")
	heads.Set("Transfer-Encoding", "chunked")
	heads.Set("Connection", "close") // probably not needed...
	heads.Set("Trailer", "X-Content-SHA256")
	heads.Set("Trailer", "X-Content-Length")

	err = w.WriteStatusLine(response.StatusOK)
	if err != nil {
		fmt.Printf("error writing status line: %v", err)
		return
	}
	err = w.WriteHeaders(heads)
	if err != nil {
		fmt.Printf("error writing headers: %v", err)
		return
	}

	chunk := make([]byte, 1024)
	var total bytes.Buffer
	for {
		n, err := resp.Body.Read(chunk)
		fmt.Printf("read %d bytes:\n%s\n", n, string(chunk[:n]))

		if n > 0 {
			// accumulate read bytes
			total.Write(chunk[:n])

			// only write however many bytes we read from current Read() call
			_, werr := w.WriteChunkedBody(chunk[:n])
			if werr != nil {
				fmt.Printf("error writing chunk: %v\n", werr)
				return
			}
		}

		if err == io.EOF {
			_, _ = w.WriteChunkedBodyDone()
			break
		}
		// actual unexpected error
		if err != nil {
			fmt.Printf("error reading response: %v", err)
			break
		}
	}

	fmt.Println("finished writing all chunks to the body :)")
	fmt.Println("ready for a new connection...")
}

func ProxyHandlerWithTrailers(w *response.Writer, req *request.Request) {
	httpBinPrefix := "/httpbin/"
	redirect := strings.HasPrefix(req.RequestLine.RequestTarget, httpBinPrefix)

	// make default response if we are not redirecting
	if !redirect {
		body := response.StatusOKBody
		heads := response.GetDefaultHeaders(len(body))
		err := heads.Update("Content-Type", "text/html")
		if err != nil {
			panic("we should always be able to update Content-Type")
		}
		server.WriteResponse(w, response.StatusOK, heads, body)
		return
	}

	// make request to httpbin to get content
	redirTarget := strings.TrimPrefix(req.RequestLine.RequestTarget, httpBinPrefix)
	resp, err := http.Get("https://httpbin.org/" + redirTarget)
	if err != nil {
		fmt.Printf("error getting response from https://httpbin.org/: %v", err)
		return
	}
	defer resp.Body.Close()

	heads := headers.NewHeaders()
	heads.Set("Content-Type", "text/plain")
	heads.Set("Transfer-Encoding", "chunked")
	heads.Set("Connection", "close") // probably not needed...
	heads.Set("Trailer", "X-Content-SHA256")
	heads.Set("Trailer", "X-Content-Length")

	err = w.WriteStatusLine(response.StatusOK)
	if err != nil {
		fmt.Printf("error writing status line: %v", err)
		return
	}
	err = w.WriteHeaders(heads)
	if err != nil {
		fmt.Printf("error writing headers: %v", err)
		return
	}

	chunk := make([]byte, 1024)
	var total bytes.Buffer
	for {
		n, err := resp.Body.Read(chunk)
		fmt.Printf("read %d bytes:\n%s\n", n, string(chunk[:n]))

		if n > 0 {
			// accumulate read bytes
			total.Write(chunk[:n])

			// only write however many bytes we read from current Read() call
			_, werr := w.WriteChunkedBody(chunk[:n])
			if werr != nil {
				fmt.Printf("error writing chunk: %v\n", werr)
				return
			}
		}

		if err == io.EOF {
			_, _ = w.WriteChunkedBodyDoneWithTrailers()
			break
		}
		// actual unexpected error
		if err != nil {
			fmt.Printf("error reading response: %v", err)
			break
		}
	}

	trails := headers.NewHeaders()
	hash := sha256.Sum256(total.Bytes())
	trails.Set("X-Content-SHA256", fmt.Sprintf("%X", hash))
	trails.Set("X-Content-Length", fmt.Sprintf("%d", total.Len()))
	err = w.WriteTrailers(trails)
	if err != nil {
		fmt.Printf("error writing trailers: %v", err)
		return
	}

	fmt.Println("finished writing all chunks to the body and the trailers :)")
	fmt.Println("ready for a new connection...")
}

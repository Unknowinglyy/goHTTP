package handlers

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
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
	if err := heads.Update("Content-Type", "text/html"); err != nil {
		panic("we should always be able to update Content-Type")
	}

	// can use this function instead for conveniency:
	// server.WriteResponse(w, status, heads, body)

	if err := w.WriteStatusLine(status); err != nil {
		fmt.Printf("error writing status line: %v", err)
		return
	}
	if err := w.WriteHeaders(heads); err != nil {
		fmt.Printf("error writing headers: %v", err)
		return
	}
	if _, err := w.WriteBody([]byte(body)); err != nil {
		fmt.Printf("error writing body: %v", err)
		return
	}
}

func ProxyHandler(w *response.Writer, req *request.Request) {
	httpBinPrefix := "/httpbin/"
	redirect := strings.HasPrefix(req.RequestLine.RequestTarget, httpBinPrefix)

	if !redirect {
		// make default response if we are not redirecting
		body := response.StatusOKBody
		heads := response.GetDefaultHeaders(len(body))
		if err := heads.Update("Content-Type", "text/html"); err != nil {
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

	if err = w.WriteStatusLine(response.StatusOK); err != nil {
		fmt.Printf("error writing status line: %v", err)
		return
	}
	if err = w.WriteHeaders(heads); err != nil {
		fmt.Printf("error writing headers: %v", err)
		return
	}

	chunk := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(chunk)
		fmt.Printf("read %d bytes:\n%s\n", n, string(chunk[:n]))

		if n > 0 {
			// only write however many bytes we read from current Read() call
			if _, werr := w.WriteChunkedBody(chunk[:n]); werr != nil {
				fmt.Printf("error writing chunk: %v\n", werr)
				return
			}
		}

		if err == io.EOF {
			break
		}

		// unexpected error
		if err != nil {
			fmt.Printf("error reading response: %v", err)
			break
		}
	}
	_, _ = w.WriteChunkedBodyDone()
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

	if err = w.WriteStatusLine(response.StatusOK); err != nil {
		fmt.Printf("error writing status line: %v", err)
		return
	}
	if err = w.WriteHeaders(heads); err != nil {
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
			if _, werr := w.WriteChunkedBody(chunk[:n]); werr != nil {
				fmt.Printf("error writing chunk: %v\n", werr)
				return
			}
		}

		if err == io.EOF {
			break
		}

		// actual unexpected error
		if err != nil {
			fmt.Printf("error reading response: %v", err)
			break
		}
	}
	_, _ = w.WriteChunkedBodyDoneWithTrailers()

	trails := headers.NewHeaders()
	hash := sha256.Sum256(total.Bytes())
	trails.Set("X-Content-SHA256", fmt.Sprintf("%X", hash))
	trails.Set("X-Content-Length", fmt.Sprintf("%d", total.Len()))
	if err = w.WriteTrailers(trails); err != nil {
		fmt.Printf("error writing trailers: %v", err)
		return
	}

	fmt.Println("finished writing all chunks to the body and the trailers :)")
	fmt.Println("ready for a new connection...")
}

func BinaryDataHandler(w *response.Writer, req *request.Request) {
	status := response.StatusOK
	body := response.StatusOKBody
	vid := req.RequestLine.RequestTarget == "/video"
	var payload []byte
	var contentType string

	if vid {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("error getting working dir: %v\n", err)
			return
		}

		payload, err = os.ReadFile(wd + "/assets/vim.mp4")
		if err != nil {
			fmt.Printf("error reading video file: %v", err)
			return
		}
		contentType = "video/mp4"
	} else {
		payload = []byte(body)
		contentType = "text/html"
	}

	heads := response.GetDefaultHeaders(len(payload))
	if err := heads.Update("Content-Type", contentType); err != nil {
		panic("failed to update Content-Type")
	}

	if err := w.WriteStatusLine(status); err != nil {
		fmt.Printf("error writing headers: %v\n", err)
	}
	if err := w.WriteHeaders(heads); err != nil {
		fmt.Printf("error writing headers: %v\n", err)
	}
	if _, err := w.WriteBody(payload); err != nil {
		fmt.Printf("error writing headers: %v\n", err)
	}
}

package response

import (
	"fmt"
	"io"
	"strconv"

	"goHttp/internal/headers"
)

type StatusCode uint16

var (
	ErrorInvalidStatus = fmt.Errorf("encountered an invalid status code")
	ErrorNoHeaders     = fmt.Errorf("found no headers to write for response")
)

const (
	StatusOK        StatusCode = 200
	StatusBad       StatusCode = 400
	StatusInServErr StatusCode = 500
	version                    = "HTTP/1.1"
)

func WriteStatusLine(w io.Writer, status StatusCode) error {
	statusLine := ""
	switch status {
	case StatusOK:
		statusLine = version + " 200 OK\r\n"
	case StatusBad:
		statusLine = version + " 400 Bad Request\r\n"
	case StatusInServErr:
		statusLine = version + " 500 Internal Server Error\r\n"
	default:
		return ErrorInvalidStatus
	}
	_, err := w.Write([]byte(statusLine))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	heads := headers.NewHeaders()
	heads.Set("Content-Length", strconv.Itoa(contentLen))
	heads.Set("Connection", "close")
	heads.Set("Content-Type", "text/plain")
	return heads
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	if len(h) == 0 {
		return ErrorNoHeaders
	}
	for key, val := range h {
		line := fmt.Sprintf("%s: %s\r\n", key, val)
		_, err := w.Write([]byte(line))
		if err != nil {
			return err
		}
	}

	// need extra CRLF to separate headers from body
	_, err := w.Write([]byte("\r\n"))
	return err
}

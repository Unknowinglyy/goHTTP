package response

import (
	"fmt"
	"io"
	"strconv"

	"goHttp/internal/headers"
	"goHttp/internal/request"
)

type (
	StatusCode  uint16
	writerState string
)

var (
	ErrorInvalidStatus        = fmt.Errorf("encountered an invalid status code")
	ErrorNoHeaders            = fmt.Errorf("found no headers to write for response")
	ErrorInvalidWriteSequence = fmt.Errorf("have not followed the correct order of response writes")
)

const (
	StatusOK     StatusCode = 200
	StatusOKBody string     = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

	StatusBad     StatusCode = 400
	StatusBadBody string     = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

	StatusInServErr     StatusCode = 500
	StatusInServErrBody string     = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

	WriteEmptyState       writerState = "haven't written anything yet"
	WriteStatusLineState  writerState = "writing status line"
	WriteHeadersState     writerState = "writing headers"
	WriteBodyState        writerState = "writing body"
	WriteChunkedBodyState writerState = "writing chunked body"
	WriteDoneState        writerState = "done writing everything"

	version = "HTTP/1.1"
)

type Writer struct {
	state writerState
	conn  io.Writer
}

func NewWriter(conn io.Writer) *Writer {
	return &Writer{state: WriteEmptyState, conn: conn}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	heads := headers.NewHeaders()
	heads.Set("Content-Length", strconv.Itoa(contentLen))
	heads.Set("Connection", "close")
	heads.Set("Content-Type", "text/plain")
	return heads
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != WriteEmptyState {
		fmt.Printf("current write state: %s", w.state)
		return ErrorInvalidWriteSequence
	}
	w.state = WriteStatusLineState

	statusLine := ""
	switch statusCode {
	case StatusOK:
		statusLine = version + " 200 OK\r\n"
	case StatusBad:
		statusLine = version + " 400 Bad Request\r\n"
	case StatusInServErr:
		statusLine = version + " 500 Internal Server Error\r\n"
	default:
		return ErrorInvalidStatus
	}

	_, err := w.conn.Write([]byte(statusLine))
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != WriteStatusLineState {
		fmt.Printf("current write state: %s", w.state)
		return ErrorInvalidWriteSequence
	}
	if len(headers) == 0 {
		return ErrorNoHeaders
	}

	w.state = WriteHeadersState

	for key, val := range headers {
		header := fmt.Sprintf("%s: %s\r\n", key, val)
		_, err := w.conn.Write([]byte(header))
		if err != nil {
			return err
		}
	}

	// need extra CRLF to separate headers from body
	_, err := w.conn.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != WriteHeadersState {
		fmt.Printf("current write state: %s", w.state)
		return 0, ErrorInvalidWriteSequence
	}
	w.state = WriteBodyState

	if len(p) == 0 {
		return 0, nil
	}

	n, err := w.conn.Write(p)
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state == WriteHeadersState {
		w.state = WriteChunkedBodyState
	}
	if w.state != WriteChunkedBodyState {
		return 0, ErrorInvalidWriteSequence
	}

	total := 0
	num := len(p)
	if num == 0 {
		return 0, nil
	}
	CRLF := request.CRLF
	sizeLine := fmt.Sprintf("%X%s", num, string(CRLF))
	chunk := [2]string{sizeLine, string(p) + string(CRLF)}
	for _, v := range chunk {
		n, err := w.conn.Write([]byte(v))
		if err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != WriteChunkedBodyState {
		return 0, ErrorInvalidWriteSequence
	}
	w.state = WriteDoneState

	total := 0
	CRLF := request.CRLF
	endingChunk := [2]string{"0" + string(CRLF), string(CRLF)}
	for _, v := range endingChunk {
		n, err := w.conn.Write([]byte(v))
		if err != nil {
			return 0, err
		}
		total += n
	}

	return total, nil
}

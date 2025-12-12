package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"

	"goHttp/internal/headers"
)

const buffSize = 8

type ParseState int

const (
	DoneState        ParseState = iota
	InitializedState            // need to parse request line
	ParsingHeadersState
	ParsingBodyState
)

var (
	// "carriage return and line feed"
	CRLF                   = []byte("\r\n")
	slash                  = []byte("/")
	space                  = []byte(" ")
	ErrorParseRequestLine  = fmt.Errorf("error when parsing the request line")
	ErrorParseDoneState    = fmt.Errorf("trying to parse data that is already parsed")
	ErrorUnknownState      = fmt.Errorf("encountered an unknown request state")
	ErrorInvalidNumParts   = fmt.Errorf("invalid number of parts in request lines")
	ErrorInvalidMethodName = fmt.Errorf("method does not contain only captial alphabetic characters")
	ErrorNoSlash           = fmt.Errorf("couldn't find '/' in HTTP version")
	ErrorUnexectedEOF      = fmt.Errorf("unexpected EOF: missing end of headers")
	ErrorBodyLengthGreater = fmt.Errorf("actual body length is greater than reported body length")
	ErrorBodyLengthLesser  = fmt.Errorf("actual body length is less than reported body length")
)

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       ParseState
}

func NewRequest() *Request {
	return &Request{state: InitializedState, Headers: headers.NewHeaders()}
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case InitializedState:
		req, n, err := parseRequestLine(data)
		if err != nil {
			return 0, errors.Join(ErrorParseRequestLine, err)
		}
		if n == 0 {
			// zero bytes were parsed with no error, simply needs more data
			return 0, nil
		}

		r.RequestLine = req.RequestLine
		r.state = ParsingHeadersState

		return n, nil
	case ParsingHeadersState:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.state = ParsingBodyState
		}
		return n, nil

	case ParsingBodyState:
		val, err := r.Headers.Get("content-length")
		if err != nil {
			return 0, err
		}

		// assuming that if no "content-length" header,
		// there is no body present so nothing to parse
		if val == "" {
			r.state = DoneState
			return 0, nil
		}

		i, err := strconv.Atoi(val)
		if err != nil {
			return 0, err
		}

		n, err := parseBody(r, data, i)
		if err != nil {
			return 0, err
		}
		return n, nil
	case DoneState:
		return 0, ErrorParseDoneState
	default:
		return 0, ErrorUnknownState
	}
}

func parseBody(req *Request, data []byte, expectedLength int) (int, error) {
	actual := len(data)
	if actual < expectedLength {
		// not enough data, request for more
		return 0, nil
	}

	if actual > expectedLength {
		return 0, ErrorBodyLengthGreater
	}

	req.Body = data[:expectedLength]
	req.state = DoneState
	return expectedLength, nil
}

func (r *Request) parse(data []byte) (int, error) {
	// bytes parsed on this run
	totalBytesParsed := 0
	for r.state != DoneState {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			// need to read in more data, returning number of bytes successfully parsed
			return totalBytesParsed, nil
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buff := make([]byte, buffSize)
	readToIndex := 0

	req := NewRequest()

	for req.state != DoneState {
		// grow buffer if it is full
		if readToIndex == len(buff) {
			temp := make([]byte, cap(buff)*2)
			// only copy read bytes that have not yet been parsed
			// (everything after readToIndex should be garbage values
			// since we have only read up until readToIndex)
			copy(temp, buff[:readToIndex])
			buff = temp
		}

		// make sure additional bytes read are appended to end of buff
		nBytes, err := reader.Read(buff[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state == ParsingBodyState {
					val, _ := req.Headers.Get("content-length")
					expect, _ := strconv.Atoi(val)

					// readToIndex holds number of unparsed bytes
					actual := readToIndex

					if actual < expect {
						return nil, ErrorBodyLengthLesser
					}
				}

				// by the time the final read goes off, we should be done parsing
				if req.state != DoneState {
					return nil, ErrorUnexectedEOF
				}
				break
			}
			return nil, err
		}
		// keeping track of how many bytes that were actually read
		readToIndex += nBytes

		// only parse the bytes in the buff that were actually read
		num, err := req.parse(buff[:readToIndex])
		if err != nil {
			return nil, err
		}
		// nothing was parsed with no error, simply need more data.
		// so we should try to read in more
		if num == 0 {
			continue
		}

		// shrinking buff by removing data from buff that was successfully parsed
		// copy starting from where the parser ended (num),
		// all the way till the bytes read but not parsed (readToIndex)
		copy(buff, buff[num:readToIndex])

		// these bytes were parsed and removed from the buff,
		// time to reset where we will be reading from
		readToIndex -= num
	}
	return req, nil
}

func parseRequestLine(data []byte) (*Request, int, error) {
	if !bytes.Contains(data, CRLF) {
		// didn't find a carriage return, so wait for more data
		return nil, 0, nil
	}
	lines := bytes.Split(data, CRLF)
	firstLine := lines[0]
	parts := bytes.Split(firstLine, space)

	if len(parts) != 3 {
		return nil, 0, ErrorInvalidNumParts
	}

	if ok := onlyUpper(parts[0]); !ok {
		return nil, 0, ErrorInvalidMethodName
	}

	idx := bytes.Index(parts[2], slash)
	if idx == -1 {
		return nil, 0, ErrorNoSlash
	}

	// should only be the supported value (1.1)
	version := parts[2][idx+1:]

	reqLine := RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HTTPVersion:   string(version),
	}
	// only consuming the first request line and the carriage return
	numBytes := len(firstLine) + len(CRLF)
	req := Request{RequestLine: reqLine}
	return &req, numBytes, nil
}

func onlyUpper(slice []byte) bool {
	// there is no "captial empty string"
	if len(slice) == 0 {
		return false
	}

	for _, char := range slice {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	return true
}

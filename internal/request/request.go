package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
)

const (
	buffSize    = 8
	INITIALIZED = 0
	DONE        = 1
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
)

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	// 0 = initialized
	// 1 = done
	state int
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case INITIALIZED:
		req, n, err := parseRequestLine(data)
		if err != nil {
			return 0, errors.Join(ErrorParseRequestLine, err)
		}
		if n == 0 {
			// err is nil here
			// zero bytes were parsed with no error, simply needs more data
			return 0, nil
		}

		r.RequestLine = req.RequestLine
		r.state = DONE

		return n, nil
	case DONE:
		return 0, ErrorParseDoneState
	default:
		return 0, ErrorUnknownState
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	// "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"
	/*
		// 0 GET / HTTP/1.1
		// 1 Host: localhost:42069
		// 2 User-Agent: curl/7.81.0
		// 3 Accept: (star/star)
		// 4 (this is the MANDATORY \r\n before the body)
		// 5 (this is the body) (might not be here)
	*/

	buff := make([]byte, buffSize)
	readToIndex := 0

	req := Request{state: 0}

	for req.state != DONE {

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
		n, err := reader.Read(buff[readToIndex:])
		if err != nil {
			return nil, err
		}
		// keeping track of how many bytes that were actually read
		readToIndex += n

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

		// removing data from buff that was successfully parsed
		// copy starting from where the parser ended (num),
		// all the way till the bytes read but not parsed (readToIndex)
		copy(buff, buff[num:readToIndex])

		// these bytes were parsed and removed from the buff,
		// time to reset where we will be reading from
		readToIndex -= num
	}
	return &req, nil
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
		log.Printf("%v is not upper", slice)
		return false
	}

	for _, char := range slice {
		if char < 'A' || char > 'Z' {
			log.Printf("%v is not upper", char)
			return false
		}
	}
	return true
}

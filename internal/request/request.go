package request

import (
	"bytes"
	"errors"
	"io"
	"log"
)

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	state       int // 0 = initialized, 1 = done
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case 0:
		// call parseRequestLine
	case 1:
		return 0, errors.New("trying to parse data in a done state")
	default:
		return 0, errors.New("encountered an unknown state")
	}
	return 0, errors.New("not implemented yet")
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

	buff, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	sep := []byte("\r\n")

	reqLine := bytes.Split(buff, sep)[0]
	request, _, err := parseRequestLine(reqLine)
	return request, err
}

func parseRequestLine(data []byte) (*Request, int, error) {
	returnChar := []byte("\r\n")
	slash := []byte("/")
	temp := bytes.Split(data, returnChar)

	if len(temp) != 3 {
		return nil, 0, errors.New("invalid number of parts in request line")
	}

	if ok := onlyUpper(temp[0]); !ok {
		return nil, 0, errors.New("method does not contain only captial alphabetic characters")
	}

	idx := bytes.Index(temp[2], slash)
	if idx == -1 {
		return nil, 0, errors.New("couldn't find '/' in HTTP version")
	}

	// should only be the supported value (1.1)
	version := temp[2][idx+1:]

	reqLine := RequestLine{Method: string(temp[0]), RequestTarget: string(temp[1]), HTTPVersion: string(version)}
	numBytes := len(temp[0]) + len(temp[1]) + len(version)
	req := Request{RequestLine: reqLine}
	return &req, numBytes, nil
}

func onlyUpper(slice []byte) bool {
	// no version of captial empty string
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

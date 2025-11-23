package request

import (
	"errors"
	"io"
	"log"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
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
		return &Request{}, err
	}

	str := string(buff)
	reqLine := strings.Split(str, "\r\n")[0]
	request, err := parseRequestLine(reqLine)
	return request, err
}

func parseRequestLine(str string) (*Request, error) {
	temp := strings.Split(str, " ")
	if len(temp) != 3 {
		return &Request{}, errors.New("invalid number of parts in request line")
	}

	if ok := onlyUpper(temp[0]); !ok {
		return &Request{}, errors.New("method does not contain only captial alphabetic characters")
	}

	idx := strings.Index(temp[2], "/")
	if idx == -1 {
		return &Request{}, errors.New("couldn't find '/' in HTTP version")
	}

	// should only be the supported value (1.1)
	version := temp[2][idx+1:]

	reqLine := RequestLine{Method: temp[0], RequestTarget: temp[1], HTTPVersion: version}
	req := Request{RequestLine: reqLine}
	return &req, nil
}

func onlyUpper(str string) bool {
	// no version of captial empty string
	if str == "" {
		log.Printf("%v is not upper", str)
		return false
	}
	for _, v := range str {
		if !unicode.IsUpper(v) {
			log.Printf("%v is not upper", v)
			return false
		}
	}
	return true
}

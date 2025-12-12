package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

const (
	colon   = ":"
	symbols = "!#$%&'*+-.^_`|~"
)

var (
	CRLF                   = []byte("\r\n")
	ErrorParseNoColon      = fmt.Errorf("found no colon while parsing header")
	ErrorNoFieldName       = fmt.Errorf("found no field name while parsing header")
	ErrorSpaceBeforeColon  = fmt.Errorf("found a space between field name and colon while parsing header")
	ErrorInvalidCharInName = fmt.Errorf("found an invalid character in the field name while parsing header")
	ErrorInvalidFieldName  = fmt.Errorf("gave an invalid field name when trying to access field value")
)

func NewHeaders() Headers {
	return make(map[string]string)
}

func (h Headers) Get(fieldName string) (string, error) {
	valid := validateFieldName(fieldName)
	if !valid {
		return "", ErrorInvalidFieldName
	}

	value, ok := h[strings.ToLower(fieldName)]
	if !ok {
		return "", nil
	}

	return value, nil
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	// "Parse will be called over and over until all headers are parsed..."
	// "can only parse one key:value pair at a time"

	endIdx := bytes.Index(data, CRLF)
	if endIdx == -1 {
		// found no CRLF, need more data
		return 0, false, nil
	}

	line := data[:endIdx]

	// empty line means no more headers to parse
	if len(line) == 0 {
		return len(CRLF), true, nil
	}

	colonIdx := bytes.Index(line, []byte(colon))
	switch colonIdx {
	case -1:
		return 0, false, ErrorParseNoColon
	case 0:
		return 0, false, ErrorNoFieldName
	}

	// check for space between colon and field name
	if unicode.IsSpace(rune(line[colonIdx-1])) {
		return 0, false, ErrorSpaceBeforeColon
	}

	// while spaces before the field name are allowed, the field name itself
	// shouldn't contain spaces
	fieldName := strings.TrimSpace(string(line[:colonIdx]))
	if fieldName == "" {
		return 0, false, ErrorNoFieldName
	}
	fieldName = strings.ToLower(fieldName)
	if ok := validateFieldName(fieldName); !ok {
		return 0, false, ErrorInvalidCharInName
	}

	fieldValue := strings.TrimSpace(string(line[colonIdx+1:]))
	// check if key already in map
	val, valid := h[fieldName]
	if valid {
		newVal := val + ", " + fieldValue
		h[fieldName] = newVal
	} else {
		h[fieldName] = fieldValue
	}

	// done is false when we get valid header line (could be more to parse)
	// Parse should be called until done is true
	return endIdx + len(CRLF), false, nil
}

func validateFieldName(s string) bool {
	for i := 0; i < len(s); i++ {
		byte := s[i]

		// should be ASCII only
		if byte > 127 {
			return false
		}
		if unicode.IsLetter(rune(byte)) || unicode.IsNumber(rune(byte)) || strings.ContainsRune(symbols, rune(byte)) {
			continue
		} else {
			return false
		}
	}
	return true
}

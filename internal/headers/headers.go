package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

var (
	Seperator                = "\r\n"
	ErrHeaderFieldMalformed  = fmt.Errorf("header field malformed")
	ErrHeaderNameMalformed   = fmt.Errorf("header field name malformed")
	ErrHeaderFieldsSeperator = fmt.Errorf("header field seperator missing")
	ErrHeaderKeyInvalidChar  = fmt.Errorf("header key contains invalid character")
	ErrKeyNotFound           = fmt.Errorf("key not found error")
)

func NewHeaders() *Headers {
	return &Headers{}
}

func isValidHeaderKey(key string) bool {
	specialCharacters := "!#$%&'*+-.^_`|~"
	for _, c := range key {
		if (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') && (c < '0' || c > '9') && !strings.Contains(specialCharacters, string(c)) {
			return false
		}
	}
	return true
}

func (h *Headers) Get(key string) (string, bool) {
	value, ok := (*h)[strings.ToLower(key)]
	if !ok {
		return "", false
	}
	return value, true
}

func (h *Headers) Parse(data []byte) (n int, done bool, err error) {
	readN := 0
	for {
		idx := bytes.Index(data[readN:], []byte(Seperator))
		if idx == -1 {
			break
		}

		headerLine, nextLine, found := bytes.Cut(data[readN:], []byte(Seperator))
		if !found {
			return readN, false, nil
		}

		if len(headerLine) == 0 {
			return readN + len(Seperator), true, nil
		}

		// Check if header name is valid
		fieldSeperatorIdx := bytes.Contains(headerLine, []byte(" :"))
		if fieldSeperatorIdx {
			return readN, false, ErrHeaderNameMalformed
		}

		fieldName, fieldValue, found := bytes.Cut(headerLine, []byte(":"))
		if !found {
			return readN, false, ErrHeaderFieldsSeperator
		}

		fieldNameStr := strings.ToLower(strings.TrimSpace(string(fieldName)))
		fieldValueStr := strings.TrimSpace(string(fieldValue))
		if !isValidHeaderKey(fieldNameStr) {
			return readN, false, ErrHeaderKeyInvalidChar
		}
		fieldValueFound, exists := (*h)[fieldNameStr]
		if exists {
			fieldValueStr = fmt.Sprintf("%s, %s", fieldValueFound, fieldValueStr)
		}
		(*h)[fieldNameStr] = fieldValueStr

		readN += idx + len(Seperator)

		if string(nextLine) == Seperator {
			return readN + len(Seperator), true, nil
		}
	}
	return readN, false, nil
}

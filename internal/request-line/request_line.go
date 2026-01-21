package requestline

import (
	"bytes"
	"fmt"
	"strings"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var Separator = "\r\n"

var (
	ErrHttpVersion      = fmt.Errorf("http format incorrect")
	ErrMethod           = fmt.Errorf("method not found")
	ErrRequestMalformed = fmt.Errorf("request malformed")
)

func Parse(data []byte) (*RequestLine, int, error) {
	requestLine, _, found := bytes.Cut(data, []byte(Separator))
	splitLine := bytes.Split(requestLine, []byte(" "))
	if !found {
		return nil, 0, nil
	}
	if len(splitLine) != 3 {
		return nil, 0, ErrRequestMalformed
	}

	httpPart := bytes.Split(splitLine[2], []byte("/"))
	if len(httpPart) != 2 {
		return nil, 0, nil
	}

	method := string(splitLine[0])
	path := string(splitLine[1])
	httpVersion := string(httpPart[1])

	if strings.ToUpper(method) != method {
		return nil, 0, ErrMethod
	}

	if httpVersion != "1.1" {
		return nil, 0, ErrHttpVersion
	}

	parsedLength := len(requestLine) + len(Separator)
	return &RequestLine{Method: method, RequestTarget: path, HttpVersion: httpVersion}, parsedLength, nil
}

package request

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"httpfromtcp/internal/headers"
	requestline "httpfromtcp/internal/request-line"
)

type (
	ParserState string
	Request     struct {
		RequestLine requestline.RequestLine
		Headers     headers.Headers
		Body        []byte
		State       ParserState
	}
)

const (
	StateInit        ParserState = "initialized"
	StateHeaderParse ParserState = "parsing headers"
	StateBodyParse   ParserState = "parsing body"
	StateDone        ParserState = "done"
	StateErr         ParserState = "error"
)

var (
	ErrRequestLineMalformed = fmt.Errorf("request-line malformed")
	ErrStream               = fmt.Errorf("error reading stream")
	ErrParsing              = fmt.Errorf("error parsing request")
	ErrBodyLengthMismatch   = fmt.Errorf("missmatch in the body length")
)

var err error

func (r *Request) parse(data []byte) (int, error) {
	read := 0
	for {
		switch r.State {
		case StateInit:
			rl, n, err := requestline.Parse(data[read:])
			if err != nil {
				r.SwitchToErrorState()
				return read, err
			}

			if n == 0 {
				return 0, nil
			}

			read += n
			r.RequestLine = *rl
			r.NextState()

		case StateHeaderParse:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				r.SwitchToErrorState()
				return read, err
			}
			read += n
			if done {
				r.NextState()
			} else {
				return read, nil
			}

		case StateBodyParse:
			length, found := r.Headers.Get("Content-Length")
			if !found {
				r.NextState()
				continue
			}

			expectedLength, _ := strconv.Atoi(length)
			bodyLength := len(data[read:])
			if bodyLength < expectedLength {
				return read, nil
			}

			r.Body = data[read : read+expectedLength]
			read += expectedLength
			r.NextState()

		case StateErr:
			return 0, err

		case StateDone:
			return read, nil
		}
	}
}

func (r *Request) NextState() {
	switch r.State {
	case StateInit:
		r.State = StateHeaderParse
	case StateHeaderParse:
		r.State = StateBodyParse
	case StateBodyParse:
		r.State = StateDone
	default:
		return
	}
}

func (r *Request) SwitchToErrorState() {
	r.State = StateErr
}

func (r *Request) Done() bool {
	return r.State == StateDone || r.State == StateErr
}

func NewRequest() *Request {
	return &Request{
		State:   StateInit,
		Headers: make(headers.Headers),
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := NewRequest()

	buf := make([]byte, 1024)
	bufIdx := 0

	for !request.Done() {
		n, err := reader.Read(buf[bufIdx:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				// Check if we're still expecting body data
				if request.State == StateBodyParse {
					if length, found := request.Headers.Get("Content-Length"); found && length != "0" {
						return nil, ErrBodyLengthMismatch
					}
				}
				request.NextState()
				break
			}
			return nil, ErrStream
		}

		bufIdx += n
		readN, err := request.parse(buf[:bufIdx])
		if err != nil {
			return nil, errors.Join(ErrParsing, err)
		}

		copy(buf, buf[readN:bufIdx])
		bufIdx -= readN
	}
	return request, nil
}

package response

import (
	"fmt"
	"io"

	"httpfromtcp/internal/headers"
)

type (
	StatusCode int
	Writer     struct {
		writer io.Writer
	}
)

type HandlerError struct {
	StatusCode StatusCode
	Message    []byte
}

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
	NotFound            StatusCode = 404
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{w}
}

func NewHandlerError(sc StatusCode, message string) *HandlerError {
	return &HandlerError{StatusCode: sc, Message: []byte(message)}
}

func (w *Writer) WriteStatusLine(sc StatusCode) error {
	var statusLine []byte
	switch sc {
	case OK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case BadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case InternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		statusLine = []byte(fmt.Appendf([]byte("HTTP/1.1 "), "%d\r\n", sc))
	}

	_, err := w.writer.Write(statusLine)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	header := headers.NewHeaders()
	(*header)["Content-Length"] = fmt.Sprint(contentLen)
	(*header)["Connection"] = "close"
	(*header)["Content-Type"] = "text/plain"
	return *header
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	b := []byte{}
	for key, value := range headers {
		b = fmt.Appendf(b, "%s: %s\r\n", key, value)
	}
	b = fmt.Appendf(b, "\r\n")
	_, err := w.writer.Write(b)
	return err
}

func (w *Writer) WriteError(e HandlerError) error {
	err := w.WriteStatusLine(BadRequest)
	if err != nil {
		return err
	}
	h := GetDefaultHeaders(0)
	err = w.WriteHeaders(h)

	return err
}

func (w *Writer) WriteBody(body []byte) error {
	_, err := w.writer.Write(body)
	return err
}

package server

import (
	"fmt"
	"log"
	"net"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/templates"
)

type ServerState string

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	State   ServerState
	handler Handler
}

const (
	InitialState ServerState = "initial"
	ClosedState  ServerState = "closed"
)

var (
	ErrAcceptConnection    = fmt.Errorf("error accepting connection")
	ErrConnectionListening = fmt.Errorf("error listening for connections")
	ErrConnectionClosed    = fmt.Errorf("connection is closed")
	ErrConnectionClosing   = fmt.Errorf("error closing connection")
)

func NewServer(handler Handler) Server {
	return Server{
		State:   InitialState,
		handler: handler,
	}
}

func (S *Server) isClosed() bool {
	return S.State == ClosedState
}

func Serve(port int, handler Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, ErrConnectionListening
	}
	server := NewServer(handler)

	go server.listen(ln)

	return &server, nil
}

func (S *Server) Close() {
	S.State = ClosedState
}

func (S *Server) listen(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		if S.isClosed() {
			return
		}
		go S.handle(conn)
	}
}

func (S *Server) handle(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println(ErrConnectionClosing, err)
		}
	}()

	w := response.NewWriter(conn)
	defer templates.Recover(w)

	request := templates.Must(request.RequestFromReader(conn))
	S.handler(w, request)
}

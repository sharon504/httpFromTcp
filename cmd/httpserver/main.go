package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port int = 42069

func handler(w *response.Writer, req *request.Request) *response.HandlerError {
	var err error
	var statusCode response.StatusCode
	var contentLength int
	var body string
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		statusCode = response.BadRequest
		contentLength = 31
		body = "Your problem is not my problem\n"
	case "/myproblem":
		statusCode = response.InternalServerError
		contentLength = 16
		body = "Woopsie, my bad\n"
	default:
		statusCode = response.NotFound
		contentLength = 15
		body = "All good, frfr\n"
	}
	err = w.WriteStatusLine(statusCode)
	if err != nil {
		return response.NewHandlerError(response.InternalServerError, err.Error())
	}
	h := response.GetDefaultHeaders(contentLength)
	err = w.WriteHeaders(h)
	if err != nil {
		return response.NewHandlerError(response.InternalServerError, err.Error())
	}
	err = w.WriteBody([]byte(body))
	if err != nil {
		return response.NewHandlerError(response.InternalServerError, err.Error())
	}
	return nil
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

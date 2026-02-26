package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"httpfromtcp/internal/templates"
)

const port int = 42069

func respond400() []byte {
	return []byte(`
	<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
	`)
}

func respond500() []byte {
	return []byte(`
	<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
	`)
}

func respond200() []byte {
	return []byte(`
	<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
	`)
}

func handler(w *response.Writer, req *request.Request) {
	var statusCode response.StatusCode
	var body []byte
	defer templates.Recover(w)
	target := req.RequestLine.RequestTarget

	switch {
	case strings.HasPrefix(target, "/yourproblem"):
		statusCode = response.BadRequest
		body = respond400()
	case strings.HasPrefix(target, "/myproblem"):
		statusCode = response.InternalServerError
		body = respond500()
	case strings.HasPrefix(target, "/httpbin/"):
		suffix := strings.TrimPrefix(target, "/httpbin/")
		res := templates.Must(http.Get("https://httpbin.org/" + suffix))
		templates.ErrorOnlyMust(w.WriteStatusLine(200))

		h := response.GetDefaultHeaders(0)
		h.Replace("Content-Type", "text/plain")
		h.Set("Transfer-Encoding", "chunked")
		h.Set("Trailer", "X-Content-SHA256")
		h.Set("Trailer", "X-Content-Length")
		h.Delete("Content-Length")
		templates.ErrorOnlyMust(w.WriteHeaders(h))

		fullBody := []byte{}
		for {
			data := make([]byte, 1024)
			n, err := res.Body.Read(data)
			if err != nil {
				break
			}
			fullBody = append(fullBody, data[:n]...)
			templates.Must(w.WriteChunkedBody(data[:n]))
		}
		templates.Must(w.WriteChunkedBodyDone())
		tailers := headers.NewHeaders()
		out := sha256.Sum256(fullBody)
		tailers.Set("X-Content-SHA256", string(out[:]))
		tailers.Set("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
		templates.ErrorOnlyMust(w.WriteHeaders(*tailers))
		return
	case strings.HasPrefix(target, "/video"):
		templates.ErrorOnlyMust(w.WriteStatusLine(200))
		data := templates.Must(os.ReadFile("assets/vim.mp4"))

		h := response.GetDefaultHeaders(len(data))
		h.Replace("Content-Type", "video/mp4")
		templates.ErrorOnlyMust(w.WriteHeaders(h))
		templates.ErrorOnlyMust(w.WriteBody(data))
	default:
		statusCode = 200
		body = respond200()
	}

	templates.ErrorOnlyMust(w.WriteStatusLine(statusCode))
	h := response.GetDefaultHeaders(len(body))
	h.Set("Content-Type", "text/html")
	templates.ErrorOnlyMust(w.WriteHeaders(h))
	templates.ErrorOnlyMust(w.WriteBody(body))
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

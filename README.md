# httpfromtcp

A custom HTTP server implementation in Go, built from scratch on top of TCP sockets.

## Overview

This project demonstrates how HTTP works at the protocol level by implementing an HTTP server that handles TCP connections directly, rather than using Go's built-in `net/http` library.

## Features

- **Custom HTTP Parser** - Parses HTTP requests from raw TCP connections
- **HTTP Response Writer** - Builds HTTP responses (status line, headers, body)
- **Chunked Transfer Encoding** - Supports HTTP chunked transfer encoding with trailers
- **HTTP Proxy** - Proxies requests to httpbin.org with trailer support
- **Static File Serving** - Serves video files with proper content types
- **Graceful Shutdown** - Handles SIGINT/SIGTERM for clean server shutdown

## Available Routes

| Route | Description |
|-------|-------------|
| `/` | Returns 200 OK with success message |
| `/yourproblem` | Returns 400 Bad Request |
| `/myproblem` | Returns 500 Internal Server Error |
| `/httpbin/*` | Proxies requests to httpbin.org with chunked encoding and trailers |
| `/video` | Serves a sample video file |

## Running

```bash
go run cmd/httpserver/main.go
```

Server starts on port 42069.

## Project Structure

```
httpfromtcp/
├── cmd/
│   └── httpserver/     # HTTP server application
├── internal/
│   ├── headers/       # HTTP header parsing and manipulation
│   ├── request/       # HTTP request parser
│   ├── response/      # HTTP response writer
│   ├── server/        # TCP server implementation
│   └── templates/     # Helper utilities
└── assets/            # Static files (video)
```

## How It Works

1. **TCP Listener** - Creates a TCP listener on the specified port
2. **Connection Handling** - Accepts TCP connections and spawns goroutines
3. **Request Parsing** - Parses raw bytes into HTTP request objects
4. **Response Building** - Writes HTTP response (status line, headers, body)
5. **Connection Cleanup** - Properly closes TCP connections

## Learning Resources

This implementation shows:
- How HTTP requests are formatted over the wire
- How chunked transfer encoding works
- How to handle HTTP trailers
- The difference between TCP and HTTP abstraction levels

package main

import (
	"fmt"
	"net"

	"httpfromtcp/internal/request"
)

func main() {
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Print("Listening error")
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Connection error")
			return
		}
		defer func() {
			if err := conn.Close(); err != nil {
				fmt.Println("Error closing connection")
			}
		}()

		fmt.Println("A connection has been accepted")
		data, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println("Error", err)
		}
		fmt.Println("Request Line:")
		fmt.Printf("- %s\n", data.RequestLine.Method)
		fmt.Printf("- %s\n", data.RequestLine.RequestTarget)
		fmt.Printf("- %s\n", data.RequestLine.HttpVersion)
		fmt.Println("Header values: ")
		for key, value := range data.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

	}
}

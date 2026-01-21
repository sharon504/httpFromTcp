package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

func getLinesChannel(conn io.ReadCloser) <-chan string {
	ch := make(chan string, 1)
	go func() {
		defer func() {
			if err := conn.Close(); err != nil {
				fmt.Println("Error closing connection")
			}
		}()
		defer close(ch)
		var currentLine string
		for {
			data := make([]byte, 8)
			n, r := conn.Read(data)
			if r != nil {
				break
			}
			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				currentLine += string(data[:i])
				data = data[i+1:]
				ch <- currentLine
				currentLine = ""
			}
			currentLine += string(data)
		}
		if len(currentLine) != 0 {
			ch <- currentLine
		}
	}()
	return ch
}

func main() {
	ln, err := net.Listen("upd", ":42069")
	if err != nil {
		fmt.Print("Listening error")
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Printf("read: %s\n", line)
		}
	}
}

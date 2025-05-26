package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

func main() {
	tcpServer, err := net.ResolveTCPAddr(TYPE, HOST+":"+PORT)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP(TYPE, nil, tcpServer)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	// writer goroutine: read from stdin and send to server
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input:", err)
				return
			}
			_, err = conn.Write([]byte(line))
			if err != nil {
				fmt.Println("Write data failed:", err)
				return
			}
			if line == "/quit\n" {
				conn.Close()
				os.Exit(0)
			}
		}
	}()

	// buffer to make data
	buffer := make([]byte, 1024)
	for {
		msg, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Server closed connection.")
			return
		}
		fmt.Print(string(buffer[:msg]))
	}
}

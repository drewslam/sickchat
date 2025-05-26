package main

import (
	"fmt"
	"log"
	"net"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

type Client struct {
	conn    net.Conn
	id      string
	manager *ClientManager
	out     chan string
}

type ClientManager struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan string
}

func (c *Client) read() {
	buffer := make([]byte, 1024)
	for {
		bytes, err := c.conn.Read(buffer)
		if err != nil {
			break
		} else {
			message := string(buffer[:bytes])
			c.manager.broadcast <- fmt.Sprintf("%s: %s", c.id, message)
		}
	}
	c.manager.unregister <- c
}

func (c *Client) write() {
	for {
		message, ok := <-c.out
		if !ok {
			break
		}
		_, err := c.conn.Write([]byte(message))
		if err != nil {
			break
		}
	}
}

func (cm *ClientManager) run() {
	for {
		select {
		case client := <-cm.register:
			cm.clients[client.id] = client
			fmt.Println("Client registered:", client.id)

		case client := <-cm.unregister:
			delete(cm.clients, client.id)
			if _, ok := cm.clients[client.id]; ok {
				close(client.out)
				client.conn.Close()
			}
			fmt.Println("Client unregistered:", client.id)

		case message := <-cm.broadcast:
			for _, client := range cm.clients {
				client.out <- message
			}
			fmt.Println("Broadcasting message:", message)
		}
	}
}

func main() {
	// start listener
	listener, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		log.Print(err)
	}
	defer listener.Close()

	// initialize and start ClientManager
	manager := &ClientManager{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan string),
	}
	go manager.run()

	// accept loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
		}

		// create new Client
		id := conn.RemoteAddr().String()
		client := &Client{
			id:      id,
			conn:    conn,
			manager: manager,
			out:     make(chan string),
		}

		// register the client
		manager.register <- client

		// start client read/write goroutines
		go client.read()
		go client.write()
	}
}

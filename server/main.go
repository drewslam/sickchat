// sickchat - A simple terminal chat client and server
// Copyright (C) 2025 Andrew Souza
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/drewslam/sickchat/common"
)

const (
	HOST = common.HOST
	PORT = common.PORT
	TYPE = common.TYPE
)

var clientIDCounter = 0
var clients = make(map[int]*Client)
var mu sync.Mutex

type Client struct {
	id      int
	conn    net.Conn
	manager *ClientManager
	out     chan string
}

type ClientManager struct {
	clients    map[int]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan string
}

func (c *Client) read() {
	reader := bufio.NewReader(c.conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line != "" {
			c.manager.broadcast <- fmt.Sprintf("%d: %s", c.id, line)
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
		if !strings.HasSuffix(message, "\n") {
			message += "\n"
		}
		_, err := c.conn.Write([]byte(message))
		if err != nil {
			fmt.Printf("Write error to client %d: %v\n", c.id, err)
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

func broadcastUserList() {
	var ids []string
	for id := range clients {
		ids = append(ids, fmt.Sprintf("%d", id))
	}
	if len(ids) == 0 {
		return
	}
	userListMsg := "USERS:" + strings.Join(ids, ",") + "\n"

	for _, client := range clients {
		client.conn.Write([]byte(userListMsg))
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
		clients:    make(map[int]*Client),
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
		clientIDCounter++
		clientID := clientIDCounter
		client := &Client{
			id:      clientID,
			conn:    conn,
			manager: manager,
			out:     make(chan string),
		}

		clients[clientID] = client

		fmt.Fprintf(conn, "ID:%d\n", clientID)

		broadcastUserList()

		// register the client
		manager.register <- client

		// start client read/write goroutines
		go client.read()
		go client.write()
	}
}

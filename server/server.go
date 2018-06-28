package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

const (
	address  = "localhost:8080"
	protocol = "tcp"
)

var (
	clients     map[net.Conn]int
	connections chan net.Conn
	messages    chan string
	mux         sync.Mutex
	clientID    int
)

func connectionHandler() {
	for {
		select {

		case connection := <-connections:
			log.Println("New client connected.")

			clients[connection] = clientID
			clientID++

			go func(connection net.Conn, clientID int) {
				reader := bufio.NewReader(connection)
				for {
					message, err := reader.ReadString('\n')
					if err != nil {
						break
					}
					messages <- fmt.Sprintf("|Client %d|: %s", clientID, message)
				}
			}(connection, clientID)

		case message := <-messages:
			for connection := range clients {
				go func(connection net.Conn, message string) {
					mux.Lock()
					defer mux.Unlock()
					_, err := connection.Write([]byte(message))
					if err != nil {
						log.Println(err)
					}
				}(connection, message)
				log.Println("New message", message)
			}
		}
	}
}

func acceptConnection(server net.Listener) {
	for {
		connection, err := server.Accept()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		connections <- connection
	}
}

func main() {
	clients = make(map[net.Conn]int)
	connections = make(chan net.Conn)
	messages = make(chan string)

	server, err := net.Listen(protocol, address)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	go acceptConnection(server)
	connectionHandler()
}

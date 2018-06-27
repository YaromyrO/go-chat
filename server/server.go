package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	address  = "localhost:8080"
	protocol = "tcp"
)

var (
	clients     map[net.Conn]string
	connections chan net.Conn
	messages    chan string
)

func connectionHandler() {
	for {
		select {

		case connection := <-connections:
			log.Println("New client connected...")

			fmt.Fprint(connection, "Enter your nickname: ")
			nickname := bufio.NewScanner(connection)
			nickname.Scan()
			clients[connection] = nickname.Text()

			go func(connection net.Conn, nickname string) {
				reader := bufio.NewReader(connection)
				for {
					message, err := reader.ReadString('\n')
					if err != nil {
						break
					}
					messages <- fmt.Sprintf("|%s|: %s", nickname, message)
				}
			}(connection, clients[connection])

		case message := <-messages:
			for connection := range clients {
				go func(connection net.Conn, message string) {
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
	clients = make(map[net.Conn]string)
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

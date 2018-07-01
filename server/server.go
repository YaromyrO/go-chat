package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

const (
	address  = "localhost:8080"
	protocol = "tcp"
)

type user struct {
	nickname string
	output   chan message
	mux sync.Mutex
}

type message struct {
	nickname string
	text     string
}

type chat struct {
	users map[string]user
	join  chan user
	leave chan user
	input chan message
}

func (chat *chat) run() {
	for {
		select {

		case user := <-chat.join:
			chat.users[user.nickname] = user
			go func() {
				chat.input <- message{
					"GO-CHAT",
					fmt.Sprintf("%s joined to GO-CHAT", user.nickname),
				}
			}()

		case user := <-chat.leave:
			delete(chat.users, user.nickname)
			go func() {
				chat.input <- message{
					"GO-CHAT",
					fmt.Sprintf("%s left from GO-CHAT", user.nickname),
				}
			}()

		case message := <-chat.input:
			for _, user := range chat.users {
				select {
				case user.output <- message:
				default:

				}
			}
		}
	}
}

func connectionHandler(connection net.Conn, chat *chat) {

	defer connection.Close()

	io.WriteString(connection, "Enter your nickname: ")
	scanner := bufio.NewScanner(connection)
	scanner.Scan()

	user := user{
		nickname: scanner.Text(),
		output: make(chan message, 10),
	}

	chat.join <- user
	defer func() {
		chat.leave <- user
	}()

	go func() {
		for scanner.Scan() {
			chat.input <- message{
				user.nickname,
				scanner.Text(),
			}
		}
	}()

	for message := range user.output {
		user.mux.Lock()
		_, err := io.WriteString(connection, message.nickname+": "+message.text+"\n")
		if err != nil {
			log.Println(err.Error())
			break
		}
		user.mux.Unlock()
	}
}

func main() {
	server, err := net.Listen(protocol, address)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer server.Close()

	chat := &chat{
		make(map[string]user),
		make(chan user),
		make(chan user),
		make(chan message),
	}

	go chat.run()

	for {
		connection, err := server.Accept()
		if err != nil {
			log.Fatalln(err.Error())
		}
		go connectionHandler(connection, chat)
	}
}

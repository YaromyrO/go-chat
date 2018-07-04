package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"flag"
)

const proto = "tcp"

type user struct {
	nickname string
	output   chan message
}

type message struct {
	nickname string
	text     string
}

type safeUsers struct {
	allUsers map[string]user
	mux      sync.Mutex
}

type chat struct {
	users safeUsers
	join  chan user
	leave chan user
	input chan message
}

func (chat *chat) run() {
	for {
		select {

		case user := <-chat.join:
			chat.users.mux.Lock()
			chat.users.allUsers[user.nickname] = user
			go func() {
				chat.input <- message{
					"GO-CHAT",
					fmt.Sprintf("%s joined to GO-CHAT", user.nickname),
				}
			}()
			chat.users.mux.Unlock()

		case user := <-chat.leave:
			chat.users.mux.Lock()
			delete(chat.users.allUsers, user.nickname)
			go func() {
				chat.input <- message{
					"GO-CHAT",
					fmt.Sprintf("%s left from GO-CHAT", user.nickname),
				}
			}()
			chat.users.mux.Unlock()

		case message := <-chat.input:
			chat.users.mux.Lock()
			for _, user := range chat.users.allUsers {
				select {
				case user.output <- message:
				}
			}
			chat.users.mux.Unlock()
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
		output:   make(chan message, 10),
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
		if message.nickname != user.nickname {
			_, err := io.WriteString(connection, message.nickname+": "+message.text+"\n")
			if err != nil {
				log.Println(err.Error())
				break
			}
		}
	}
}

func main() {
	port := flag.String("port", "8080", "server port")
	host := flag.String("host", "localhost", "server host")

	server, err := net.Listen(proto, *host + ":" + *port)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer server.Close()

	chat := &chat{
		safeUsers{allUsers: make(map[string]user)},
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

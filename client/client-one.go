package main

import (
	"net"
	"fmt"
	"bufio"
	"os"
)

func listenResponse(connection net.Conn)  {
	for {
		response := make([]byte, 1024)
		n, _ := connection.Read(response)
		fmt.Print(string(response[:n]))
	}
}

func main() {

	connection, _ := net.Dial("tcp", "127.0.0.1:8080")

	go listenResponse(connection)

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		connection.Write([]byte(text))
	}
}

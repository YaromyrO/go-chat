package main

import (
	"net"
	"bufio"
	"os"
	"fmt"
)

func main() {

	connection, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go func() {
		for {
			response := make([]byte, 1024)
			n, err := connection.Read(response)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Print(string(response[:n]))
		}
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		connection.Write([]byte(text))
	}
}
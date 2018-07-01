package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
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
			if err == io.EOF {
				fmt.Println("Server stopped !")
				fmt.Println("Exit...")
				os.Exit(1)
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

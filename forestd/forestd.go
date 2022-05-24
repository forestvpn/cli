package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	address := "localhost:2405"
	listener, _ := net.Listen("tcp", address)

	for {

		fmt.Println("Listening on " + address)

		conn, err := listener.Accept()

		fmt.Println("Incoming connection from", conn.RemoteAddr())

		if err != nil {
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	for {
		scanner := bufio.NewScanner(conn)
		data := scanner.Text()

		if err := scanner.Err(); err != nil {
			_, err := bufio.NewWriter(conn).WriteString(err.Error())

			if err != nil {
				fmt.Println(err)
			}

			break
		}

		fmt.Println(data)

	}
}

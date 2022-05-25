// Forest daemon is a wireguard controller designed to run in root.
package main

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strings"
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

// A connection handler routine.
// Sets wireguard connection down or up.
// If an error occurs it writes it to the client.
func handleClient(conn net.Conn) {
	defer conn.Close()

	for {
		scanner := bufio.NewScanner(conn)
		data := scanner.Text()

		if err := scanner.Err(); err != nil {
			respond(err.Error(), conn)
			break
		}

		request := strings.Fields(data)
		action, config := request[0], request[1]

		if action == "connect" {
			cmd := connect(config)
			stdout, err := cmd.Output()

			if err != nil {
				respond(err.Error(), conn)
			} else {
				respond(string(stdout), conn)
			}
		} else if action == "disconnect" {
			cmd := disconnect()
			stdout, err := cmd.Output()

			if err != nil {
				respond(err.Error(), conn)
			} else {
				respond(string(stdout), conn)
			}
		} else {
			respond("Not implemented", conn)
		}
	}
}

// Executes the wireguard binary to establish the connection.
func connect(config string) *exec.Cmd {
	return exec.Command("wg-quick", "up", config)
}

// Executes the wireguard binary to terminate the connection.
func disconnect() *exec.Cmd {
	return exec.Command("wg-quick", "down")
}

// A basic function to send a text message to other site.
func respond(message string, conn net.Conn) {
	_, err := bufio.NewWriter(conn).WriteString(message)

	if err != nil {
		fmt.Println(err)
	}
}

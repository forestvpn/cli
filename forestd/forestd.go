package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/getsentry/sentry-go"
)

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		TracesSampleRate: 1.0,
	})

	if err != nil {
		sentry.Logger.Panicf("sentry.Init: %s", err)
	}

	address := "localhost:9999"
	listener, err := net.Listen("tcp", address)

	if err != nil {
		sentry.CaptureException(err)
		log.Panic(err)
	}

	for {
		log.Printf("Listening on %s", address)
		conn, err := listener.Accept()

		if err != nil {
			log.Print(err)
			sentry.CaptureException(err)
			continue
		}

		log.Printf("Incoming connection from %s", conn.RemoteAddr())

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	remoteAddr := conn.RemoteAddr()
	var command *exec.Cmd
	var knownActions []string
	var status int
	var config string
	defer conn.Close()

	for {
		content, err := Read(conn, DELIMITER)

		if err != nil {
			log.Print(err)
			sentry.CaptureException(err)
		}

		if content == QUIT_SIGN {
			log.Printf("%s disconnected", remoteAddr)
			break
		}

		knownActions = append(knownActions, "connect", "disconnect")
		request := strings.Fields(content)
		action := request[0]

		log.Printf(`Incoming request "%s" from %s`, action, remoteAddr)

		if len(request) > 1 && strings.Contains(strings.Join(knownActions, ""), action) {
			config = request[1]

			log.Printf("Corresponding method found: %s", action)

			if action == knownActions[0] {
				command = exec.Command("wg-quick", "up", config)
			} else if action == knownActions[1] {
				command = exec.Command("wg-quick", "down", config)
			}

			log.Printf("Executing: %s", command.String())

			status = execute(command)
		} else if action == "status" {
			status = isActiveWireGuard()
		} else {
			status = -1
		}

		log.Printf("Responding %c to %s", status, remoteAddr)

		response := fmt.Sprintf("%c%c", status, DELIMITER)
		_, err = Write(conn, response)

		if err != nil {
			log.Print(err)
			sentry.CaptureException(err)
		}
	}
}

// Indicates status of current wireguard connection
//
// Returns:
//
// - 0 - if not connected to any wireguard peer
//
// - 1 - if connected
func isActiveWireGuard() int {
	stdout, err := exec.Command("wg", "show").Output()

	if err != nil {
		log.Print(err)
		sentry.CaptureException(err)
	}

	if len(stdout) > 0 {
		return 1
	}
	return 0
}

// Executes shell commands
// Used to start/stop wireguard connection
// Returns an exit status of a shell command executed
func execute(command *exec.Cmd) int {
	if err := command.Start(); err != nil {
		log.Print(err)
		sentry.CaptureException(err)
	}

	if err := command.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus()
			}
		} else {
			log.Print(err)
			sentry.CaptureException(err)
		}
	}
	return 0
}

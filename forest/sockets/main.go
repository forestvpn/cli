package sockets

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/forestvpn/cli/auth"
)

const (
	DELIMITER byte = '\n'
	QUIT_SIGN      = "quit!"
)

func read(conn net.Conn, delim byte) ([]byte, error) {
	reader := bufio.NewReader(conn)
	var buffer bytes.Buffer
	for {
		ba, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return []byte(""), err
		}
		buffer.Write(ba)
		if !isPrefix {
			break
		}
	}
	return buffer.Bytes(), nil
}

func write(conn net.Conn, content string) (int, error) {
	writer := bufio.NewWriter(conn)
	number, err := writer.WriteString(content)
	if err == nil {
		err = writer.Flush()
	}
	return number, err
}

func Communicate(request string) (int, error) {
	conn, err := net.DialTimeout("tcp", "localhost:9999", time.Millisecond*200)

	if err != nil {
		return 0, err
	}

	num, err := write(conn, request)

	if err != nil {
		return num, err
	}

	response, err := read(conn, DELIMITER)

	if err != nil {
		return len(response), err
	}

	request = fmt.Sprintf("%s%c", QUIT_SIGN, DELIMITER)
	num, err = write(conn, request)

	if err != nil {
		return num, err
	}

	return int(response[0]), nil
}

func Disconnect() error {
	isActive, err := IsActiveConnection()

	if err != nil {
		return err
	}

	if isActive {
		request := fmt.Sprintf("disconnect %s%c", auth.WireguardConfig, DELIMITER)
		status, err := Communicate(request)

		if err != nil {
			return err
		}

		if status != 0 {
			return fmt.Errorf(`forestd could not perform action "disconnect" (exit status: %d)`, status)
		}
	}
	return nil
}

func IsActiveConnection() (bool, error) {
	request := fmt.Sprintf("status%c", DELIMITER)
	status, err := Communicate(request)

	if err != nil {
		return false, err
	}

	if status > 0 {
		return true, nil
	}
	return false, nil
}

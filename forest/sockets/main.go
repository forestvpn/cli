package sockets

import (
	"fmt"
	"forest/utils"
	"net"
	"time"
)

func Communicate(request string) (int, error) {
	conn, err := net.DialTimeout("tcp", "localhost:9999", time.Millisecond*200)

	if err != nil {
		return 0, err
	}

	num, err := Write(conn, request)

	if err != nil {
		return num, err
	}

	response, err := Read(conn, DELIMITER)

	if err != nil {
		return len(response), err
	}

	request = fmt.Sprintf("%s%c", QUIT_SIGN, DELIMITER)
	num, err = Write(conn, request)

	if err != nil {
		return num, err
	}

	return int(response[0]), nil
}

func Disconnect() error {
	request := fmt.Sprintf("status%c", DELIMITER)
	status, err := Communicate(request)

	if err != nil {
		return err
	}

	if status > 0 {
		request := fmt.Sprintf("disconnect %s%c", utils.WireguardConfig, DELIMITER)
		status, err = Communicate(request)

		if err != nil {
			return err
		}

		if status != 0 {
			return fmt.Errorf(`forestd could not perform action "disconnect" (exit status: %d)`, status)
		}
	}
	return nil
}

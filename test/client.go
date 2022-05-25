// The client to test forestd
package client

import (
	"bufio"
	"net"
)

var forestdAddress = "localhost:2405"

func sendRequest(request string) (written int, err error) {
	conn, _ := net.Dial("tcp", forestdAddress)
	defer conn.Close()
	return bufio.NewWriter(conn).WriteString(request)
}

func getResponse() {

}

package utils

import (
	"errors"
	"net"
	"os/exec"
	"strings"

	externalip "github.com/glendc/go-external-ip"
)

func GetDefaultGateway() (net.IP, error) {
	stdout, err := exec.Command("ip", "route").Output()

	if err != nil {
		return nil, err
	}

	for _, record := range strings.Split(string(stdout), "\n") {
		if strings.Contains(record, "default") {
			return net.ParseIP(strings.Join(strings.Split(record, " ")[1:], " ")), nil
		}
	}
	return nil, errors.New("default route not found")
}

func GetHostIP() (net.IP, error) {
	return externalip.DefaultConsensus(nil, nil).ExternalIP()
}

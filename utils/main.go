package utils

import (
	"net"
	"os/exec"
	"strings"

	externalip "github.com/glendc/go-external-ip"
)

func GetDefaultGateway() string {
	stdout, _ := exec.Command("ip", "route").Output()

	for _, record := range strings.Split(string(stdout), "\n") {
		if strings.Contains(record, "default") {
			return strings.Join(strings.Split(record, " ")[1:], " ")
		}
	}

	return ""
}

func GetHostIP() (net.IP, error) {
	return externalip.DefaultConsensus(nil, nil).ExternalIP()
}

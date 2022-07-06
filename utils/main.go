package utils

import (
	"os/exec"
	"strings"
)

func GetExistingRoutes() ([]string, error) {
	stdout, err := exec.Command("ip", "r").Output()

	if err != nil {
		return nil, err
	}

	var addr []string

	for _, s := range strings.Split(string(stdout), "\n") {
		if !strings.Contains(s, "default") {
			addr = append(addr, strings.Split(s, " ")[0])
		}
	}

	return addr, nil

}

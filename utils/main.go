package utils

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/forestvpn/cli/auth"
	externalip "github.com/glendc/go-external-ip"
	"github.com/go-resty/resty/v2"
)

func ip2Net(ip string) string {
	return strings.Join(strings.Split(ip, ".")[:3], ".") + ".0/24"
}

func getExistingRoutes() (map[int]string, error) {
	existingRoutes := make(map[int]string)
	stdout, _ := exec.Command("ip", "route").Output()

	for index, record := range strings.Split(string(stdout), "\n") {
		if !strings.Contains(record, "default") && len(record) > 0 {
			target := strings.Split(record, " ")[0]

			_, network, err := net.ParseCIDR(target)

			if err != nil {
				ip := net.ParseIP(target)

				if ip == nil {
					return existingRoutes, fmt.Errorf("error parsing routing table network: %s", ip)
				}

				_, network, err = net.ParseCIDR(ip2Net(ip.String()))

				if err != nil {
					return existingRoutes, err
				}
			}

			existingRoutes[index] = network.String()

		}
	}

	return existingRoutes, nil
}

func GetHostIP() (net.IP, error) {
	return externalip.DefaultConsensus(nil, nil).ExternalIP()
}

func GetAllowedIps() (*resty.Response, error) {
	url := "https://hooks.arcemtene.com/wireguard/allowedips"
	existingRoutes, err := getExistingRoutes()

	if err != nil {
		return nil, err
	}

	var disallowed []string

	for _, value := range existingRoutes {
		disallowed = append(disallowed, value)
	}

	param := strings.Join(disallowed, ",")
	fmt.Println(param)

	return auth.Client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"allowed":    "0.0.0.0/0",
			"disallowed": param,
		}).
		Get(url)

}

package utils

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	forestvpn_api "github.com/forestvpn/api-client-go"
	"github.com/forestvpn/cli/auth"
	externalip "github.com/glendc/go-external-ip"
	"github.com/go-resty/resty/v2"
)

func ip2Net(ip string) string {
	return strings.Join(strings.Split(ip, ".")[:3], ".") + ".0/24"
}

func getExistingRoutes() ([]string, error) {
	var existingRoutes []string
	stdout, _ := exec.Command("ip", "route").Output()

	for _, record := range strings.Split(string(stdout), "\n") {
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

			existingRoutes = append(existingRoutes, network.String())

		}
	}

	hostip, err := getHostIP()

	if err != nil {
		return existingRoutes, err
	}

	ipnet := ip2Net(hostip.String())
	existingRoutes = append(existingRoutes, ipnet)

	return existingRoutes, nil
}

func getHostIP() (net.IP, error) {
	return externalip.DefaultConsensus(nil, nil).ExternalIP()
}

func GetAllowedIps(peer forestvpn_api.WireGuardPeer) (*resty.Response, error) {
	url := "https://hooks.arcemtene.com/wireguard/allowedips"
	existingRoutes, err := getExistingRoutes()

	if err != nil {
		return nil, err
	}

	activeSShClientIps, err := getActiveSshClientIps()

	if err != nil {
		return nil, err
	}

	disallowed := strings.Join(append(existingRoutes, activeSShClientIps...), ",")
	allowed := strings.Join(peer.GetAllowedIps(), ",")

	fmt.Println(disallowed)

	return auth.Client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"allowed":    allowed,
			"disallowed": disallowed,
		}).
		Get(url)

}

func getActiveSshClientIps() ([]string, error) {
	out, err := exec.Command("who").Output()

	if err != nil {
		return nil, err
	}

	records := strings.Split(string(out), "\n")
	ips := make([]string, len(records)-1)

	for i, record := range records {
		if len(record) > 1 {
			record := strings.ReplaceAll(record, " ", " ")
			ip := strings.Split(record, " ")[4]
			ip = strings.Replace(ip, "(", "", 1)
			ip = strings.Replace(ip, ")", "", 1)

			if net.ParseIP(ip) != nil {
				ips[i] = ip2Net(ip)
			}
		}
	}
	return ips, err
}

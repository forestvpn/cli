package utils

import (
	"fmt"
	"math/big"
	"net"
	"os/exec"
	"strings"

	"github.com/c-robinson/iplib"
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

func GetAllowedIpsLocal(peer forestvpn_api.WireGuardPeer) ([]string, error) {
	existingRoutes, err := getExistingRoutes()

	if err != nil {
		return nil, err
	}

	activeSShClientIps, err := getActiveSshClientIps()

	if err != nil {
		return nil, err
	}

	disallowed := append(existingRoutes, activeSShClientIps...)
	allowed := peer.GetAllowedIps()
	var allowednew []string

	for _, dnet := range disallowed {

		dnet4 := iplib.Net4FromStr(dnet)

		if dnet4.Count() == 0 {
			dnet6 := iplib.Net6FromStr(dnet)

			if dnet6.Count() == big.NewInt(0) {
				break
			}

			allowednew = append(allowednew, dnet)
			break
		}

		for _, anet := range allowed {
			anet4 := iplib.Net4FromStr(anet)

			if anet4.Count() == 0 {
				break
			}

			for anet4.ContainsNet(dnet4) {
				asubnets, err := anet4.Subnet(0)

				if err != nil {
					return nil, err
				}

				for _, asubnet := range asubnets {
					if !asubnet.ContainsNet(dnet4) {
						allowednew = append(allowednew, asubnet.String())
					} else {
						anet4 = asubnet
					}
				}
			}
			allowednew = append(allowednew, anet4.String())
		}
	}
	return allowednew, nil

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
	var ips []string

	for _, record := range records {
		if len(record) > 1 {
			ip := record[strings.Index(record, "(")+1 : strings.Index(record, ")")]

			if net.ParseIP(ip) != nil {
				ips = append(ips, ip2Net(ip))
			}
		}
	}

	return ips, err
}

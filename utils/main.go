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

func excludeIPv4NetFromIPv4Net(from iplib.Net4, target iplib.Net4) ([]string, error) {
	var allowedips []string

	for from.String() != target.String() {
		if !from.ContainsNet(target) {
			allowedips = append(allowedips, from.String())
			break
		}

		asubnets, err := from.Subnet(0)

		if err != nil {
			return allowedips, err
		}

		for _, asubnet := range asubnets {
			if !asubnet.ContainsNet(target) {
				allowedips = append(allowedips, asubnet.String())
			} else {
				from = asubnet
			}
		}
	}

	return allowedips, nil

}

func excludeIPv4NetFromIPv6Net(from iplib.Net6, target iplib.Net4) ([]string, error) {
	var allowedips []string

	for from.String() != target.String() {
		if !from.ContainsNet(target) {
			allowedips = append(allowedips, from.String())
			break
		}

		asubnets, err := from.Subnet(0, 0)

		if err != nil {
			return allowedips, err
		}

		for _, asubnet := range asubnets {
			if !asubnet.ContainsNet(target) {
				allowedips = append(allowedips, asubnet.String())
			} else {
				from = asubnet
			}
		}
	}

	return allowedips, nil

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
	var allowednew = make(map[int]string)
	var allowedips []string
	var anet6 iplib.Net6

	for _, dnet := range disallowed {
		dnet4 := iplib.Net4FromStr(dnet)

		for _, anet := range allowed {
			anet4 := iplib.Net4FromStr(anet)

			if anet4.Count() == 1 {
				anet6 = iplib.Net6FromStr(anet)

				if anet6.Count() == big.NewInt(1) {
					break
				}
			}

			if anet4.Count() > 1 {
				allowedips, err = excludeIPv4NetFromIPv4Net(anet4, dnet4)
			} else {
				allowedips, err = excludeIPv4NetFromIPv6Net(anet6, dnet4)
			}

			if err != nil {
				return allowedips, err
			}

			for i, n := range allowedips {
				_, ok := allowednew[i]

				if !ok {
					allowednew[i] = n

				}
			}

		}
	}
	allowedips = make([]string, len(allowednew))

	for _, value := range allowednew {
		allowedips = append(allowedips, value)
	}

	return allowedips, nil

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

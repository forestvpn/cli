package utils

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/c-robinson/iplib"
	forestvpn_api "github.com/forestvpn/api-client-go"
	externalip "github.com/glendc/go-external-ip"
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

func GetAllowedIps(peer forestvpn_api.WireGuardPeer) ([]string, error) {
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
	var parsednets []iplib.Net

	for _, network := range allowed {
		_, net, err := iplib.ParseCIDR(network)

		if err != nil {
			return allowed, err
		}

		parsednets = append(parsednets, net)

	}

	for _, net := range parsednets {
		for _, d := range disallowed {
			_, network, err := iplib.ParseCIDR(d)

			if err != nil {
				break
			}

			if !net.ContainsNet(network) {
				allowednew = append(allowednew, net.String())
			} else {
				ipv4net := iplib.Net4FromStr(net.String())

				if ipv4net.Count() == 1 {
					allowednew = append(allowednew, ipv4net.String())
				} else {
					for ipv4net.String() != network.String() {

						subnets, err := ipv4net.Subnet(0)

						if err != nil {
							break
						}

						for _, subnet := range subnets {
							if subnet.ContainsNet(network) {
								ipv4net = subnet
							} else {
								allowednew = append(allowednew, subnet.String())
							}
						}

					}
				}
			}

			// out := fmt.Sprintf("%s: %t %s", net.String(), net.ContainsNet(network), network.String())
			// fmt.Println(out)
		}
	}
	return allowednew, nil

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

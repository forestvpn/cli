// utils is a package with network related utility functions.

package utils

import (
	"net"
	"os/exec"
	"strings"

	"github.com/c-robinson/iplib"
	externalip "github.com/glendc/go-external-ip"
)

// ip2Net is a function for converting an IP address value, e.g. 127.0.0.1, into a network with mask of 24 bits, e.g. 127.0.0.0/24.
func ip2Net(ip string) string {
	return strings.Join(strings.Split(ip, ".")[:3], ".") + ".0/24"
}

// GetExistingRoutes  is a function that calls a netstat shell command to get system routing table.
// Then it extracts all the IP addresses from the shell command's output.
// Each of an IP addresses is converted into related network using ip2Net in order to be compitable with Wireguard configuration format.
// Returns a slice of a networks (as a strings) representing system routing table.
func GetExistingRoutes() ([]string, error) {
	var existingRoutesMap = make(map[string]bool)
	var existingRoutes []string

	stdout, _ := exec.Command("netstat", "-n", "-r", "-f", "inet").Output()

	for _, record := range strings.Split(string(stdout), "\n") {
		dest := strings.Split(record, " ")[0]

		_, network, err := net.ParseCIDR(dest)

		if err != nil {
			ip := net.ParseIP(dest)

			if ip != nil {
				_, network, err = net.ParseCIDR(ip2Net(ip.String()))
			}

		}

		if err == nil {
			existingRoutesMap[network.String()] = true
		}

	}

	hostip, err := getHostIP()

	if err != nil {
		return nil, err
	}

	ipnet := ip2Net(hostip.String())
	existingRoutesMap[ipnet] = true

	for k := range existingRoutesMap {
		existingRoutes = append(existingRoutes, k)
	}
	return existingRoutes, nil
}

// getHostIP returns the host's public IP address.
//
// See https://github.com/glendc/go-external-ip for more information.
func getHostIP() (net.IP, error) {
	return externalip.DefaultConsensus(nil, nil).ExternalIP()
}

// ExcludeDisallowedIps is a function that expects two slices of a network values, e.g. [127.0.0.0/8,], where disallowed is a slice of networks to be excluded from the allowed slice.
// Returns a new slice of networks formed out of the allowed slice without networks of disallowed slice.
func ExcludeDisallowedIps(allowed []string, disallowed []string) ([]string, error) {
	var netmap = make(map[string]bool)
	var allowednew []string

	for _, a := range allowed {
		containsDisallowedNetwork := false
		_, allowedNetwork, err := iplib.ParseCIDR(a)

		if err != nil {
			return nil, err
		}

		for _, d := range disallowed {
			_, disallowedNetwork, err := iplib.ParseCIDR(d)

			if err != nil {
				return nil, err
			}

			contains := allowedNetwork.ContainsNet(disallowedNetwork)

			if contains {
				if !containsDisallowedNetwork {
					containsDisallowedNetwork = true
				}

				ipv4net := iplib.Net4FromStr(allowedNetwork.String())

				if ipv4net.Count() > 1 {
					for ipv4net.String() != disallowedNetwork.String() {

						subnets, err := ipv4net.Subnet(0)

						if err != nil {
							return nil, err
						}

						for _, subnet := range subnets {
							if subnet.ContainsNet(disallowedNetwork) {
								ipv4net = subnet
							} else {
								netmap[subnet.String()] = contains
							}
						}

					}
				}
			}
		}

		if !containsDisallowedNetwork {
			netmap[a] = containsDisallowedNetwork
		}

	}

	for k := range netmap {
		_, net, err := iplib.ParseCIDR(k)

		if err != nil {
			return nil, err
		}

		resultingNetwork := iplib.Net4FromStr(net.String())

		for _, d := range disallowed {
			disallowedNetwork := iplib.Net4FromStr(d)

			if resultingNetwork.ContainsNet(disallowedNetwork) {
				delete(netmap, resultingNetwork.String())
			}
		}
	}

	for k := range netmap {
		allowednew = append(allowednew, k)
	}

	return allowednew, nil

}

// GetActiveSshClientIps is a function that calls the "who" shell command to get active ssh sessions.
// Then it extracts all the IP addresses from the command output and converts them into networks using ip2Net for a compability with Wiregaurd configuration format.
// Returns a slice of networks representing the public networks of active ssh clients.
func GetActiveSshClientIps() ([]string, error) {
	out, err := exec.Command("who").Output()

	if err != nil {
		return nil, err
	}

	records := strings.Split(string(out), "\n")
	var ips []string

	for _, record := range records {
		if strings.Count(record, "(")+strings.Count(record, ")") > 0 {
			ip := record[strings.Index(record, "(")+1 : strings.Index(record, ")")]

			if net.ParseIP(ip) != nil {
				ips = append(ips, ip2Net(ip))
			}
		}
	}

	return ips, err
}

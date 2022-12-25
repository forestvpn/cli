// utils is a package with network related utility functions.

package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/c-robinson/iplib"
	"github.com/hashicorp/go-retryablehttp"
)

var Verbose bool

const Os = runtime.GOOS

// FirebaseApiKey is stored in an environment variable and assigned during the build with ldflags.
const FirebaseApiKey = "AIzaSyArN6RVqftrSVBrEI9ZF2DiiA7gJOdkfeM"

// ApiHost is a hostname of Forest VPN back-end API that is stored in an environment variable and assigned during the build with ldflags.
const ApiHost = "api.forestvpn.com"

var InfoLogger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lmsgprefix)

// ip2Net is a function for converting an IP address value, e.g. 127.0.0.1, into a network with mask of 24 bits, e.g. 127.0.0.0/24.
func ip2Net(ip string) string {
	return strings.Join(strings.Split(ip, ".")[:3], ".") + ".0/24"
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
func GetActiveSshClients() ([]string, error) {
	var sshConnections []string
	var ips []string

	out, err := exec.Command("netstat").Output()

	if err != nil {
		return nil, err
	}

	records := strings.Split(string(out), "\n")

	for _, record := range records {
		if strings.Contains(record, "ssh") {
			sshConnections = append(sshConnections, record)
		}
	}

	for _, record := range sshConnections {
		space := regexp.MustCompile(`\s+`)
		record := space.ReplaceAllString(record, " ")
		ip := strings.Split(strings.Split(record, " ")[4], ":")[0]

		if net.ParseIP(ip) != nil {
			ips = append(ips, ip2Net(ip))
		} else {
			_ips, err := net.LookupIP(ip)

			if err == nil {
				for _, _ip := range _ips {
					ips = append(ips, ip2Net(_ip.String()))
				}
			}
		}
	}

	return ips, err
}

// humanizeDuration humanizes time.Duration output to a meaningful value,
// golang's default “time.Duration“ output is badly formatted and unreadable.
func HumanizeDuration(duration time.Duration) string {
	if duration.Seconds() < 60.0 {
		return fmt.Sprintf("%d seconds", int64(duration.Seconds()))
	}
	if duration.Minutes() < 60.0 {
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%d minutes %d seconds", int64(duration.Minutes()), int64(remainingSeconds))
	}
	if duration.Hours() < 24.0 {
		remainingMinutes := math.Mod(duration.Minutes(), 60)
		remainingSeconds := math.Mod(duration.Seconds(), 60)
		return fmt.Sprintf("%d hours %d minutes %d seconds",
			int64(duration.Hours()), int64(remainingMinutes), int64(remainingSeconds))
	}
	remainingHours := math.Mod(duration.Hours(), 24)
	remainingMinutes := math.Mod(duration.Minutes(), 60)
	remainingSeconds := math.Mod(duration.Seconds(), 60)
	return fmt.Sprintf("%d days %d hours %d minutes %d seconds",
		int64(duration.Hours()/24), int64(remainingHours),
		int64(remainingMinutes), int64(remainingSeconds))
}

func GetLocalTimezone() (string, error) {
	b, err := ioutil.ReadFile("/etc/timezone")

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(b)), nil
}

// GetHttpClient is a factory function to get http client with provided retries number.
func GetHttpClient(retries int) *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retries
	retryClient.Logger = nil
	return retryClient.StandardClient()
}

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/getsentry/sentry-go"
)

func Commit() error {
	err := exec.Command("uci", "commit", "network").Run()

	if err != nil {
		return err
	}

	return exec.Command("/etc/init.d/network", "restart").Run()
}

// IsOpenWRT is a function to determine whether cli is running on OpenWRT device.
func IsOpenWRT() bool {
	data, err := os.ReadFile("/etc/banner")

	if err != nil {
		return false
	}

	if strings.Contains(string(data), "OpenWrt") {
		return true
	}

	return false

}

func Firewall(wiregaurdInterface string) error {
	err := exec.Command("uci", "rename", "firewall.@zone[0]=lan").Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "rename", "firewall.@zone[1]=wan").Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "del_list", fmt.Sprintf("firewall.wan.network=%s", wiregaurdInterface)).Run()

	if err != nil {
		sentry.CaptureException(err)
	}

	err = exec.Command("uci", "add_list", fmt.Sprintf("firewall.wan.network=%s", wiregaurdInterface)).Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "commit", "firewall").Run()

	if err != nil {
		return err
	}

	return exec.Command("/etc/init.d/firewall", "restart").Run()
}

func Network(
	wiregaurdInterface string,
	wireguardPrivateKey string,
	wiregaurdAddresses []string,
	wiregaurdPublicKey string,
	wiregaurdPreSharedKey string,
	wiregaurdEndpointHost string,
	wiregaurdEndpointPort string,
	wireguardAllowedIps []string) error {
	err := exec.Command("uci", "delete", fmt.Sprintf("network.%s", wiregaurdInterface)).Run()

	if err != nil {
		sentry.CaptureException(err)

		if Verbose {
			InfoLogger.Println(err)
		}
	}

	err = exec.Command("uci", "set", fmt.Sprintf("network.%s=interface", wiregaurdInterface)).Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "set", fmt.Sprintf("network.%s.proto=wireguard", wiregaurdInterface)).Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "set", fmt.Sprintf("network.%s.private_key=%s", wiregaurdInterface, wireguardPrivateKey)).Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "add_list", fmt.Sprintf("network.%s.addresses=%s", wiregaurdInterface, strings.Join(wiregaurdAddresses, ","))).Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "delete", "network.wgserver").Run()

	if err != nil {
		sentry.CaptureException(err)

		if Verbose {
			InfoLogger.Println(err)
		}
	}

	err = exec.Command("uci", "set", fmt.Sprintf("network.wgserver=wireguard_%s", wiregaurdInterface)).Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "set", fmt.Sprintf("network.wgserver.public_key=%s", wiregaurdPublicKey)).Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "set", fmt.Sprintf("network.wgserver.preshared_key=%s", wiregaurdPreSharedKey)).Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "set", fmt.Sprintf("network.wgserver.endpoint_host=%s", wiregaurdEndpointHost)).Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "set", fmt.Sprintf("network.wgserver.endpoint_port=%s", wiregaurdEndpointPort)).Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "set", "network.wgserver.route_allowed_ips=1").Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "set", "network.wgserver.persistent_keepalive=25").Run()

	if err != nil {
		return err
	}

	err = exec.Command("uci", "set", fmt.Sprintf("network.wgserver.allowed_ips=%s", strings.Join(wireguardAllowedIps, ","))).Run()

	if err != nil {
		return err
	}

	return Commit()

}

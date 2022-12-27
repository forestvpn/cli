package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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
	cmd := exec.Command("uci", "rename", `firewall.@zone[0]="lan"`)

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "rename", `firewall.@zone[1]="wan"`)

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "del_list", fmt.Sprintf(`firewall.wan.network="%s"`, wiregaurdInterface))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "add_list", fmt.Sprintf(`firewall.wan.network="%s"`, wiregaurdInterface))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "commit", "firewall")

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("/etc/init.d/firewall", "restart")
	return cmd.Run()
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
	cmd := exec.Command("uci", "-q", "delete", fmt.Sprintf("network.%s", wiregaurdInterface))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.%s="interface"`, wiregaurdInterface))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.%s.proto="wireguard"`, wiregaurdInterface))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.%s.private_key="%s"`, wiregaurdInterface, wireguardPrivateKey))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "add_list", fmt.Sprintf(`network.%s.addresses="%s"`, wiregaurdInterface, strings.Join(wiregaurdAddresses, ",")))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "-q", "delete", "network.wgserver")

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.wgserver="wireguard_%s"`, wiregaurdInterface))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.wgserver.public_key="%s"`, wiregaurdPublicKey))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.wgserver.preshared_key="%s"`, wiregaurdPreSharedKey))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.wgserver.endpoint_host="%s"`, wiregaurdEndpointHost))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.wgserver.endpoint_port="%s"`, wiregaurdEndpointPort))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", `network.wgserver.route_allowed_ips="1"`)

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", `network.wgserver.persistent_keepalive="25"`)

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.wgserver.allowed_ips="%s"`, strings.Join(wireguardAllowedIps, ",")))

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("uci", "commit", "network")

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("/etc/init.d/network", "restart")
	return cmd.Run()
}

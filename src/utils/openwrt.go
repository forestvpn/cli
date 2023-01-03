package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func handleCommand(command *exec.Cmd) error {
	out, err := command.Output()

	if err != nil {
		return err
	}

	if Verbose {
		InfoLogger.Println(string(out))
	}

	return err
}

func Commit() error {
	cmd := exec.Command("uci", "commit", "network")
	err := handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("/etc/init.d/network", "restart")
	return handleCommand(cmd)
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
	err := handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.%s="interface"`, wiregaurdInterface))
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.%s.proto="wireguard"`, wiregaurdInterface))
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.%s.private_key="%s"`, wiregaurdInterface, wireguardPrivateKey))
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "add_list", fmt.Sprintf(`network.%s.addresses="%s"`, wiregaurdInterface, strings.Join(wiregaurdAddresses, ",")))
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "-q", "delete", "network.wgserver")
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.wgserver="wireguard_%s"`, wiregaurdInterface))
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.wgserver.public_key="%s"`, wiregaurdPublicKey))
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.wgserver.preshared_key="%s"`, wiregaurdPreSharedKey))
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.wgserver.endpoint_host="%s"`, wiregaurdEndpointHost))
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.wgserver.endpoint_port="%s"`, wiregaurdEndpointPort))
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", `network.wgserver.route_allowed_ips="1"`)
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", `network.wgserver.persistent_keepalive="25"`)
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	cmd = exec.Command("uci", "set", fmt.Sprintf(`network.wgserver.allowed_ips="%s"`, strings.Join(wireguardAllowedIps, ",")))
	err = handleCommand(cmd)

	if err != nil {
		return err
	}

	return Commit()

}

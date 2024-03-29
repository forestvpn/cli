// state is a package containing a structure to control Wireguard connection.

package actions

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/forestvpn/cli/auth"
	"github.com/forestvpn/cli/utils"
)

// State is a structure representing Wireguard connection state.
type State struct {
	status             bool
	WiregaurdInterface string
}

// Deprecated: setStatus is used to set a status of Wireguard connection on the State structure.
// It calls a 'wg show' shell command and analyzes it's output.
//
// Using api.ApiClientWrapper.GetStatus instead
func (s *State) setStatus() {
	s.status = false
	if utils.IsOpenWRT() {
		stdout, _ := exec.Command("uci", "show").CombinedOutput()

		if strings.Contains(string(stdout), "wgserver") {
			s.status = true
		}
	} else {
		stdout, _ := exec.Command("wg", "show").CombinedOutput()

		if len(stdout) > 0 {
			s.status = true
		}
	}
}

// GetStatus is a method to get the status of a Wireguard connection.
//
// Using api.ApiClientWrapper.GetStatus instead
func (s *State) GetStatus() bool {
	s.setStatus()
	return s.status
}

// SetUp is a method used to establish a Wireguard connection.
// It executes 'wg-quick' shell command.
func (s *State) SetUp(user_id auth.ProfileID, persist bool) error {
	var allowedIPs []string
	path := auth.ProfilesDir + string(user_id) + auth.WireguardConfig

	if utils.Os == "windows" {
		return exec.Command("wireguard", "/installtunnelservice", path).Run()
	} else if utils.IsOpenWRT() {
		device, err := auth.LoadDevice(user_id)
		if err != nil {
			return err
		}

		IPs := device.GetIps()
		if persist {
			err = utils.Firewall(s.WiregaurdInterface)
			if err != nil {
				return err
			}

			peer := device.Wireguard.GetPeers()[0]
			endpoint := strings.Split(peer.GetEndpoint(), ":")
			activeSShClient := utils.GetActiveSshClient()
			if err != nil {
				return err
			}

			if len(activeSShClient) > 0 {
				allowedIPs, err = utils.ExcludeDisallowedIps(peer.GetAllowedIps(), activeSShClient)
				if err != nil {
					return err
				}
			}

			return utils.Network(s.WiregaurdInterface, device.Wireguard.GetPrivKey(), IPs, peer.GetPubKey(), peer.GetPsKey(), endpoint[0], endpoint[1], allowedIPs)
		} else {
			err := exec.Command("ip", "link", "add", "dev", s.WiregaurdInterface, "type", "wireguard").Run()
			if err != nil {
				return err
			}

			err = exec.Command("ip", "address", "add", "dev", s.WiregaurdInterface, IPs[1]).Run()
			if err != nil {
				return err
			}

			err = exec.Command("ip", "-6", "address", "add", "dev", s.WiregaurdInterface, IPs[2]).Run()
			if err != nil {
				return err
			}

			err = exec.Command("wg", "setconf", s.WiregaurdInterface, path).Run()
			if err != nil {
				return err
			}

			err = exec.Command("ip", "link", "set", "up", "dev", s.WiregaurdInterface).Run()
			if err != nil {
				return err
			}

			return exec.Command("ip", "route", "add", "default", "dev", s.WiregaurdInterface).Run()
		}
	} else {
		return exec.Command("wg-quick", "up", path).Run()
	}
}

// SetDown is used to terminate a Wireguard connection.
// It executes 'wg-quick' shell command.
func (s *State) SetDown(user_id auth.ProfileID) error {
	configPath := auth.ProfilesDir + string(user_id) + auth.WireguardConfig
	var command *exec.Cmd
	switch {
	case utils.Os == "windows":
		command = exec.Command("wireguard", "/uninstalltunnelservice", s.WiregaurdInterface)
	case utils.IsOpenWRT():
		if err := exec.Command("uci", "-q", "delete", fmt.Sprintf("network.%s", s.WiregaurdInterface)).Run(); err != nil {
			return err
		}
		if err := exec.Command("uci", "-q", "delete", "network.wgserver").Run(); err != nil {
			return err
		}
		return utils.Commit()
	default:
		command = exec.Command("wg-quick", "down", configPath)
	}
	return command.Run()
}

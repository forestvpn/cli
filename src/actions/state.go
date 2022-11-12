// state is a package containing a structure to control Wireguard connection.

package actions

import (
	"os/exec"

	"github.com/forestvpn/cli/auth"
	"github.com/forestvpn/cli/utils"
)

// State is a structure representing Wireguard connection state.
type State struct {
	status bool
}

// Deprecated: setStatus is used to set a status of Wireguard connection on the State structure.
// It calls a 'wg show' shell command and analyzes it's output.
//
// Using api.ApiClientWrapper.GetStatus instead
func (s *State) setStatus() {
	s.status = false
	stdout, _ := exec.Command("wg", "show").CombinedOutput()

	if len(stdout) > 0 {
		s.status = true
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
func (s *State) SetUp(user_id string) error {
	var command *exec.Cmd
	path := auth.ProfilesDir + user_id + auth.WireguardConfig

	if utils.Os == "windows" {
		command = exec.Command("wireguard", "/installtunnelservice", path)
	} else {
		command = exec.Command("wg-quick", "up", path)
	}

	return command.Run()
}

// SetDown is used to terminate a Wireguard connection.
// It executes 'wg-quick' shell command.
func (s *State) SetDown(user_id string) error {
	var command *exec.Cmd
	path := auth.ProfilesDir + user_id + auth.WireguardConfig

	if utils.Os == "windows" {
		command = exec.Command("wireguard", "/uninstalltunnelservice", path)
	} else {
		command = exec.Command("wg-quick", "down", path)
	}
	return command.Run()
}

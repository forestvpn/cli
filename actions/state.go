// state is a containing a structure to control Wireguard connection.

package actions

import (
	"os/exec"
)

type State struct {
	status bool
}

// setStatus is used to set a status of Wireguard connection on the State structure.
// It calls a 'wg show' shell command and analyzes it's output.
func (s *State) setStatus() {
	s.status = false
	stdout, _ := exec.Command("wg", "show").CombinedOutput()

	if len(stdout) > 0 {
		s.status = true
	}
}

// GetStatus is a public function to get the status of a Wireguard connection.
func (s *State) GetStatus() bool {
	s.setStatus()
	return s.status
}

// SetUp is used to establish a Wireguard connection.
// It executes 'wg-quick'shell command.
func (s *State) SetUp(config string) error {
	return exec.Command("wg-quick", "up", config).Run()
}

// SetDown is used to terminate a Wireguard connection.
// It executes 'wg-quick' shell command.
func (s *State) SetDown(config string) error {
	return exec.Command("wg-quick", "down", config).Run()
}

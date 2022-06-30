package actions

import (
	"os/exec"
)

type State struct {
	status bool
}

func (s *State) setStatus() {
	s.status = false
	stdout, _ := exec.Command("wg", "show").CombinedOutput()

	if len(stdout) > 0 {
		s.status = true
	}
}

func (s *State) GetStatus() bool {
	s.setStatus()
	return s.status
}

func (s *State) SetUp(config string) error {
	return exec.Command("wg-quick", "up", config).Run()
}

func (s *State) SetDown(config string) error {
	return exec.Command("wg-quick", "down", config).Run()
}

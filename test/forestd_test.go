// Forest daemon tests
package main

import (
	"client"
	"fmt"
	"os/exec"
	"testing"
)

// The Process ID of a daemon process to control it
var pid int

func TestConnect(t testing.T) {
	setUpDaemon()
	request = "connect" + " " + ""
	written, err := sendRequest(action, config)
	setDownDaemon()

	if written != len(request) {
		t.Errorf("Bytes written during request is %d; want %d", written, len(request))
	}

	if err != nil {
		t.Errorf(err.Error())
	}

}

func TestDisconnect() {
	setUpDaemon()
	written, err := sendRequest("disconnect", "")
	setDownDaemon()
}

// Sets Forestd up before testing
func setUpDaemon() {
	if err := exec.Command("go", "run", "forestd/forestd.go").Start(); err != nil {
		fmt.Println("Unable to start a daemon")
	} else {
		pid = cmd.Process.Pid
	}
}

// Sets Forestd down after testing
func setDownDaemon() {
	if err := exec.Command("kill", pid).Start(); err != nil {
		fmt.Println("Error stopping the daemon")
	}
}

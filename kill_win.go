// +build windows

package main

import (
	"os"
	"os/exec"
	"strconv"
)

func killCmd(config *Config) {
	if config.cmd == nil {
		config.logToTextarea("[kcptun] not running")
		return
	}

	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(config.cmd.Process.Pid))
	kill.Stderr = os.Stderr
	kill.Stdout = os.Stdout
	if err := kill.Run(); err != nil {
		config.logToTextarea("kill cmd: " + err.Error())
		return
	}
	config.cmd = nil
}

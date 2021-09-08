// +build !windows

package main

func killCmd(config *Config) {
	err := config.cmd.Process.Kill()
	if err != nil {
		config.logToTextarea("failed to kill: " + string(config.cmd.Process.Pid))
		return
	}

	config.logToTextarea("killed")
	return
}

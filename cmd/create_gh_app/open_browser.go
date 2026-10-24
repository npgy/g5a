package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func openBrowser(url string) error {
	var command string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		command = "open"
	case "linux":
		command = "xdg-open"
	case "windows":
		command = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	args = append(args, url)
	return exec.Command(command, args...).Start()
}

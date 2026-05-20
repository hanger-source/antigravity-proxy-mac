package main

import (
	"os/exec"
)

func runOsascript(script string) {
	exec.Command("osascript", "-e", script).Run()
}

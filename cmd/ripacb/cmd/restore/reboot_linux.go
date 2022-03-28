package restore

import (
	"os/exec"
)

func Reboot() {
	exec.Command("reboot").Run()
}

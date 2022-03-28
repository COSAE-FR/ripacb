//go:build freebsd && !pfsense
// +build freebsd,!pfsense

package restore

import "os/exec"

func Reboot() {
	exec.Command("reboot").Run()
}

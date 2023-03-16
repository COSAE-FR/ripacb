//go:build freebsd && pfsense
// +build freebsd,pfsense

package restore

import (
	log "github.com/sirupsen/logrus"
	"os/exec"
)

func Reboot() {
	if err := exec.Command("/etc/rc.stop_packages").Run(); err != nil {
		log.Printf("Cannot stop packages: %s", err)
	}
	cmd := exec.Command("/etc/rc.reboot", "-r")
	if err := cmd.Start(); err == nil {
		cmd.Process.Release()
	} else {
		log.Printf("Cannot reboot: %s", err)
	}
}

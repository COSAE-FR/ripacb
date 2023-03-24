//go:build freebsd && pfsense
// +build freebsd,pfsense

package restore

import (
	log "github.com/sirupsen/logrus"
	"os/exec"
)

func Reboot() {
	log.Printf("Stopping packages...")
	if err := exec.Command("/etc/rc.stop_packages").Run(); err != nil {
		log.Printf("Cannot stop packages: %s", err)
	}
	log.Printf("Launching rc.reboot...")
	if err := exec.Command("/etc/rc.reboot").Run(); err != nil {
		log.Printf("Cannot run rc.reboot: %s", err)
	}
}

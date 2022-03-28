//go:build freebsd && !pfsense
// +build freebsd,!pfsense

package restore

func Reboot() {
	exec.Command("reboot").Run()
}

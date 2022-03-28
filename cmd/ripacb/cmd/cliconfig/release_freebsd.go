package cliconfig

import (
	"github.com/COSAE-FR/riputils/common"
	"github.com/ashcrow/osrelease"
	"io/ioutil"
	"strings"
)

const (
	platformFile = "/etc/platform"
	versionFile  = "/etc/version"
)

func OSRelease(platform, version string) (string, string) {
	var err error
	var or osrelease.OSRelease

	or, err = osrelease.New()
	if err == nil {
		platform = or.NAME
		version = or.VERSION_ID
	}
	if common.FileExists(platformFile) {
		raw, err := ioutil.ReadFile(platformFile)
		if err == nil {
			platform = strings.TrimSpace(string(raw))
		}
	}
	if common.FileExists(versionFile) {
		raw, err := ioutil.ReadFile(versionFile)
		if err == nil {
			version = strings.TrimSpace(string(raw))
		}
	}
	return platform, version
}

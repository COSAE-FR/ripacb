//go:build !freebsd
// +build !freebsd

package cliconfig

import (
	"github.com/ashcrow/osrelease"
)

func OSRelease(platform, version string) (string, string) {
	or, err := osrelease.New()
	if err != nil {
		return platform, version
	}
	return or.NAME, or.VERSION_ID
}

package cliconfig

import (
	"encoding/xml"
	"github.com/COSAE-FR/riputils/common"
	"github.com/COSAE-FR/riputils/pfsense/configuration"
	"github.com/go-playground/validator/v10"
	"io/ioutil"
	"os"
	"regexp"
	"time"
)

const DefaultPfSenseXML = "/conf/config.xml"
const PfSenseGlobalsFile = "/etc/inc/globals.inc"

var PfSenseXML = DefaultPfSenseXML

var DefaultServerURL = "https://acb.reseau.rip"
var DefaultPassword string
var DefaultHostname = "ripsense"

var DefaultPlatform = "ripSense"
var DefaultPlatformVersion = "22.2"

var DefaultDialTimeout = 2 * time.Second
var DefaultRequestTimeout = 10 * time.Second

type Configuration struct {
	ServerURL      string `validate:"url"`
	Password       string
	Hostname       string
	DialTimeout    time.Duration
	RequestTimeout time.Duration
}

func (c Configuration) Validate() error {
	return validate.Struct(c)
}

var Config Configuration
var validate *validator.Validate

var pfConfiguration *configuration.BaseConfiguration

func getDefaultServerURL() string {
	defaultServer := DefaultServerURL
	if pfConfiguration != nil {
		if pfConfiguration.System.ACB.Server != "" {
			return pfConfiguration.System.ACB.Server
		}
	}
	if common.FileExists(PfSenseGlobalsFile) {
		globals, err := ioutil.ReadFile(PfSenseGlobalsFile)
		if err == nil {
			r, err := regexp.Compile(`"default_acb_server"\s*=>\s*"(http[^"]+)"`)
			if err == nil {
				match := r.FindStringSubmatch(string(globals))
				if len(match) == 2 {
					return match[1]
				}
			}
		}
	}
	return defaultServer
}

func getDefaultSPassword() string {
	defaultPassword := DefaultPassword
	if pfConfiguration != nil {
		if pfConfiguration.System.ACB.Password != "" {
			return pfConfiguration.System.ACB.Password
		}
	}
	return defaultPassword
}

func getDefaultHostname() string {
	defaultHostname := DefaultHostname
	if pfConfiguration != nil {
		if pfConfiguration.System.Hostname != "" {
			return pfConfiguration.System.Hostname
		}
	}
	hs, err := os.Hostname()
	if err == nil && hs != "" {
		return hs
	}
	return defaultHostname
}

func readPfSenseConfiguration() {
	if common.FileExists(PfSenseXML) {
		data, err := ioutil.ReadFile(PfSenseXML)
		if err != nil {
			return
		}
		err = xml.Unmarshal(data, &pfConfiguration)
		if err != nil {
			return
		}
	}

}

func init() {
	validate = validator.New()
	readPfSenseConfiguration()
	DefaultServerURL = getDefaultServerURL()
	DefaultPassword = getDefaultSPassword()
	DefaultHostname = getDefaultHostname()
	DefaultPlatform, DefaultPlatformVersion = OSRelease(DefaultPlatform, DefaultPlatformVersion)
}

var SkipConfig = false

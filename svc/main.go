package main

import (
	"github.com/COSAE-FR/ripacb/pkg/acb"
	"github.com/COSAE-FR/ripacb/pkg/utils"
	"github.com/COSAE-FR/riputils/svc"
	"github.com/COSAE-FR/riputils/svc/shared"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func ConfigParser(logger *log.Entry, config shared.Config) (svc.Daemonizer, error) {
	log.SetLevel(log.InfoLevel)
	cfg := &acb.ServiceConfig{}
	file, err := ioutil.ReadFile(config.Conf)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.Check(); err != nil {
		return nil, err
	}
	return acb.NewService(logger, cfg)
}

func main() {
	svc.StartService(utils.Name+"d", ConfigParser, svc.WithDefaultConfigurationPath("/etc/acbd.yaml"), svc.WithVersion(utils.Version))
}

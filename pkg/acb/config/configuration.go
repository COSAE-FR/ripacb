package config

import (
	"github.com/COSAE-FR/riputils/tls"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
)

type Features struct {
	// Allow saving new configuration revision
	AllowSave bool `yaml:"allow_save" json:"allow_save"`
	// Allow saving new configuration revision for new devices
	AllowNew bool `yaml:"allow_new" json:"allow_new"`
	// Allow deleting revision
	AllowDelete bool `yaml:"allow_delete" json:"allow_delete"`
	// Allow restoring user saved revisions
	AllowRestoreUser bool `yaml:"allow_restore_user" json:"allow_restore_user"`
	// Maximum revision count to keep
	MaxBackups int `yaml:"max_backups" validate:"gt=0" default:"100" json:"max_backups"`
	// This server handles only portal requests
	IsPortal bool `yaml:"is_portal" json:"is_portal,omitempty"`
}

type Config struct {
	IPAddress         string             `yaml:"ip" validate:"ip4_addr,required" default:"0.0.0.0"`
	Port              uint16             `yaml:"port" validate:"gt=0,lte=65535" default:"80"`
	RequestsPerMinute int64              `yaml:"rate" default:"60"`
	Features          Features           `yaml:"features"`
	TLS               *tls.Configuration `yaml:"tls,omitempty"`
}

func (c *Config) Check() error {
	if err := defaults.Set(c); err != nil {
		return err
	}
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return err
	}
	if c.TLS != nil {
		if err := c.TLS.Check(); err != nil {
			return err
		}
	}
	return nil
}

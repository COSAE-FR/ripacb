package acb

import (
	"fmt"
	"github.com/COSAE-FR/ripacb/pkg/acb/config"
	"github.com/COSAE-FR/ripacb/pkg/acb/store"
	"github.com/COSAE-FR/ripacb/pkg/utils"
	"github.com/COSAE-FR/riputils/common/logging"
	"github.com/COSAE-FR/riputils/db"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

type StoreConfig struct {
	Type     store.Type        `yaml:"type" validate:"required" default:"directory"`
	Path     string            `yaml:"path" validate:"required_if=Type directory"`
	Database *db.Configuration `yaml:"db"`
}

func (c *StoreConfig) Check() error {
	if err := defaults.Set(c); err != nil {
		return err
	}
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return err
	}
	if c.Type == store.DBStoreType {
		return c.Database.Check()
	}
	return nil
}

func (c *StoreConfig) GetStore(logger *log.Entry) (store.Store, error) {
	switch c.Type {
	case store.DirectoryStoreType:
		return store.NewDirectoryStore(c.Path, logger)
	case store.DBStoreType:
		return store.NewDBStore(c.Database, logger)
	default:
		return nil, fmt.Errorf("unknown store type")
	}
}

type ServiceConfig struct {
	Store           StoreConfig     `yaml:"store" validate:"required"`
	Servers         []config.Config `yaml:"servers"`
	*logging.Config `yaml:"logging"`
}

func (c *ServiceConfig) Check() error {
	if err := defaults.Set(c); err != nil {
		return err
	}
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return err
	}
	for _, s := range c.Servers {
		if err := s.Check(); err != nil {
			return err
		}
	}
	return c.Store.Check()
}

type Service struct {
	store   store.Store
	log     *log.Entry
	config  *ServiceConfig
	servers []*Server
}

func NewService(entry *log.Entry, config *ServiceConfig) (*Service, error) {
	if config == nil {
		return nil, fmt.Errorf("empty configuration")
	}
	logger := config.SetupLog(utils.Name, utils.Version)
	if logger != nil {
		entry = logger
	}
	if err := config.Check(); err != nil {
		return nil, err
	}
	serviceStore, err := config.Store.GetStore(entry.WithField("component", "store"))
	if err != nil {
		return nil, err
	}
	return &Service{
		store:   serviceStore,
		log:     entry.WithField("component", "acb_service"),
		config:  config,
		servers: nil,
	}, nil
}

func (s *Service) Configure() error {
	var errorList []error
	if s.servers == nil {
		var servers []*Server
		for _, server := range s.config.Servers {
			serverConfig := server
			s, err := NewServer(s.log, s.store, &serverConfig)
			if err != nil {
				errorList = append(errorList, err)
			}
			servers = append(servers, s)
		}
		if len(errorList) > 0 {
			return fmt.Errorf("error while creating servers: %v", errorList)
		}
		s.servers = servers
	}
	for _, server := range s.servers {
		s.log.Tracef("Configuring server on %s:%d", server.config.IPAddress, server.config.Port)
		if err := server.Configure(); err != nil {
			errorList = append(errorList, err)
		}
	}
	if len(errorList) == 0 {
		return nil
	}
	return fmt.Errorf("error while configuring servers: %v", errorList)
}

func (s *Service) Start() error {
	var errorList []error
	var startedList []*Server
	for _, server := range s.servers {
		if err := server.Start(); err != nil {
			errorList = append(errorList, err)
		} else {
			startedList = append(startedList, server)
		}
	}
	if len(errorList) == 0 {
		return nil
	}
	for _, server := range startedList {
		_ = server.Stop()
	}
	return fmt.Errorf("error while starting servers: %v", errorList)
}

func (s *Service) Stop() error {
	var errorList []error
	for _, server := range s.servers {
		if err := server.Stop(); err != nil {
			errorList = append(errorList, err)
		}
	}
	if len(errorList) == 0 {
		return nil
	}
	return fmt.Errorf("error while stopping servers: %v", errorList)
}

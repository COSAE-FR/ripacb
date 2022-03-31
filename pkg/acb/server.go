package acb

import (
	"context"
	"fmt"
	config2 "github.com/COSAE-FR/ripacb/pkg/acb/config"
	"github.com/COSAE-FR/riputils/gin/ginlog"
	"github.com/COSAE-FR/riputils/gin/rate"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type Server struct {
	log      *logrus.Entry
	config   *config2.Config
	server   *http.Server
	listener net.Listener
	started  bool
	store    Store
	useTLS   bool
}

func NewServer(log *logrus.Entry, store Store, config *config2.Config) (*Server, error) {
	if store == nil {
		return nil, fmt.Errorf("store is mandatory")
	}
	tlsConfig, err := config.TLS.Generate()
	if err != nil {
		return nil, err
	}
	router := gin.New()
	_ = router.SetTrustedProxies(config.TrustedProxies)
	srv := &Server{
		server: &http.Server{
			Handler:   router,
			TLSConfig: tlsConfig,
		},
		log:    log.WithField("component", "acb_server"),
		config: config,
		store:  store,
	}
	if tlsConfig != nil {
		srv.useTLS = true
	}
	lm := rate.NewRateLimiter(time.Minute, config.RequestsPerMinute, func(c *gin.Context) (string, error) {
		return c.ClientIP(), nil
	})
	router.Use(ginlog.Logger(srv.log), lm.Middleware(), gin.Recovery())
	router.POST("/getbkp", srv.ListBackupsHandler)
	router.POST("/list", srv.ListBackupsHandler)
	router.GET("/getbkp", srv.Ping)
	router.POST("/save", srv.SaveBackupHandler)
	router.POST("/rmbkp", srv.DeleteBackupHandler)
	router.POST("/api/v1/backups", srv.ListBackupsHandler)
	router.GET("/api/v1/status", srv.Ping)
	router.POST("/api/v1/backups/new", srv.SaveBackupHandler)
	router.POST("/api/v1/backups/delete", srv.DeleteBackupHandler)

	return srv, nil
}

func (s *Server) Configure() error {
	var err error
	s.listener, err = net.Listen("tcp4", fmt.Sprintf("%s:%d", s.config.IPAddress, s.config.Port))
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Start() error {
	if s.started {
		if err := s.Stop(); err != nil {
			s.log.Errorf("stopping ACB server: %s", err)
		}
	}
	if s.listener == nil {
		if err := s.Configure(); err != nil {
			s.log.Errorf("configuring ACB server: %s", err)
			return err
		}
	}
	s.started = true
	go func() {
		if s.useTLS {
			_ = s.server.ServeTLS(s.listener, "", "")
		} else {
			_ = s.server.Serve(s.listener)
		}

	}()
	return nil
}

func (s *Server) Stop() error {
	if s.started {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			cancel()
			s.started = false
		}()
		return s.server.Shutdown(ctx)
	}
	return nil
}

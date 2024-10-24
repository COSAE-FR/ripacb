package acb

import (
	"errors"
	"github.com/COSAE-FR/ripacb/pkg/acb/bindings"
	"github.com/COSAE-FR/ripacb/pkg/acb/entity"
	"github.com/COSAE-FR/ripacb/pkg/acb/store"
	"github.com/COSAE-FR/ripacb/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"io"
	"net/http"
	"strings"
	"time"
)

// ListBackupsHandler implements the ACB /getbkp
func (s *Server) ListBackupsHandler(c *gin.Context) {
	request := &bindings.GetBackupRequest{}
	err := c.Bind(request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, &bindings.StatusResponse{
			Code:    http.StatusNotFound,
			Message: "No revision",
		})
		return
	}
	if request.DeviceKey == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, &bindings.StatusResponse{
			Code:    http.StatusBadRequest,
			Message: "No revision",
		})
		return
	}
	logger := s.log.WithField("device", request.DeviceKey)
	revisions, err := s.store.GetRevisionsForDevice(request.DeviceKey, s.config.Features)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, &bindings.StatusResponse{
			Code:    http.StatusNotFound,
			Message: "No revision",
		})
		logger.Tracef("No revision: %s", err)
		return
	}
	if request.Revision != "" {
		logger = logger.WithField("requested_revision", request.Revision)
		revision, found := revisions[request.Revision]
		if !found {
			c.AbortWithStatusJSON(http.StatusNotFound, &bindings.StatusResponse{
				Code:    http.StatusNotFound,
				Message: "No revision",
			})
			logger.Tracef("Unknown revision")
			return
		}
		if c.Request.URL.Path == "/api/v1/backups" {
			c.JSON(http.StatusOK, revision)
			return
		}
		c.Header("Content-Type", "application/vnd+rip-acb-revision+text")
		c.String(http.StatusOK, revision.MarshallText())
		return
	}
	if c.Request.URL.Path == "/api/v1/backups" {
		c.JSON(http.StatusOK, revisions)
		return
	}
	c.Header("Content-Type", "application/vnd+rip-acb-list+text")
	c.String(http.StatusOK, revisions.MarshallText())
}

func (s *Server) Ping(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusOK, &bindings.StatusResponse{
		Code:     http.StatusOK,
		Message:  "Server OK",
		Version:  utils.Version,
		Features: &s.config.Features,
	})
}

func (s *Server) SaveBackupHandler(c *gin.Context) {
	if !s.config.Features.AllowSave {
		s.log.Info("save disabled")
		c.AbortWithStatusJSON(http.StatusCreated, &bindings.StatusResponse{
			Code:    http.StatusCreated,
			Message: "Backup created",
		})
		return
	}
	request := &bindings.SaveBackupRequest{}
	err := c.Bind(request)
	if err != nil {
		s.log.Errorf("cannot parse save request: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, &bindings.StatusResponse{
			Code:    http.StatusInternalServerError,
			Message: "Invalid request",
		})
		return
	}
	ct, err := request.Content.Open()
	if err != nil {
		s.log.Errorf("cannot open attached configuration: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, &bindings.StatusResponse{
			Code:    http.StatusInternalServerError,
			Message: "Invalid request",
		})
		return
	}
	buf := new(strings.Builder)
	_, err = io.Copy(buf, ct)
	if err != nil {
		s.log.Errorf("cannot read attached configuration: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, &bindings.StatusResponse{
			Code:    http.StatusInternalServerError,
			Message: "Invalid request",
		})
		return
	}
	user, comment := entity.ParseReason(request.Reason)
	backup := &entity.Revision{
		Hash:       request.Hash,
		Content:    buf.String(),
		Reason:     request.Reason,
		Username:   user,
		Comment:    comment,
		Date:       time.Now(),
		FromPortal: false,
		Device:     request.DeviceKey,
	}
	if s.config.Features.IsPortal && request.Date != nil {
		backup.Date = *request.Date
	}
	backup.Revision = backup.Date.UTC().Format(time.RFC3339)
	validate := validator.New()
	if err := validate.Struct(backup); err != nil {
		s.log.Errorf("Invalid request: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, &bindings.StatusResponse{
			Code:    http.StatusInternalServerError,
			Message: "Invalid request",
		})
		return
	}
	if err := s.store.SetRevision(backup, s.config.Features); err != nil {
		if errors.Is(err, store.ErrMaskedToClient) {
			s.log.Errorf("cannot save revision: %s, not reporting to client", err)
			c.AbortWithStatusJSON(http.StatusCreated, &bindings.StatusResponse{
				Code:    http.StatusCreated,
				Message: "Backup created",
			})
			return
		}
		s.log.Errorf("cannot save revision: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, &bindings.StatusResponse{
			Code:    http.StatusInternalServerError,
			Message: "Invalid request",
		})
		return
	}
	c.AbortWithStatusJSON(http.StatusCreated, &bindings.StatusResponse{
		Code:    http.StatusCreated,
		Message: "Backup created",
	})
}

func (s *Server) DeleteBackupHandler(c *gin.Context) {
	if !s.config.Features.AllowDelete {
		s.log.Info("delete disabled")
		c.AbortWithStatusJSON(http.StatusCreated, &bindings.StatusResponse{
			Code:    http.StatusForbidden,
			Message: "Backup not deleted",
		})
		return
	}
	request := &bindings.GetBackupRequest{}
	err := c.Bind(request)
	if err != nil {
		s.log.Errorf("cannot parse save request: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, &bindings.StatusResponse{
			Code:    http.StatusInternalServerError,
			Message: "Invalid request",
		})
		return
	}
	if request.Revision != "" {
		if err := s.store.DeleteRevision(request.DeviceKey, request.Revision, s.config.Features); err != nil {
			if errors.Is(err, store.ErrMaskedToClient) {
				s.log.Errorf("cannot delete revision: %s, not reporting to client", err)
				c.AbortWithStatusJSON(http.StatusNotFound, &bindings.StatusResponse{
					Code:    http.StatusNotFound,
					Message: "No revision",
				})
				return
			}
			s.log.Errorf("cannot delete revision: %s", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, &bindings.StatusResponse{
				Code:    http.StatusInternalServerError,
				Message: "Cannot delete revision",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusAccepted, &bindings.StatusResponse{
			Code:    http.StatusAccepted,
			Message: "Backup deleted",
		})
		return
	}
	s.log.Errorf("No revision in request")
	c.AbortWithStatusJSON(http.StatusNotFound, &bindings.StatusResponse{
		Code:    http.StatusNotFound,
		Message: "No revision",
	})
}

package bindings

import (
	"github.com/COSAE-FR/ripacb/pkg/acb/config"
	"mime/multipart"
	"time"
)

type GetBackupRequest struct {
	Uid       string `form:"uid"`
	Version   string `form:"version"`
	Revision  string `form:"revision"`
	DeviceKey string `form:"userkey"`
}

type StatusResponse struct {
	Code     int              `json:"code"`
	Message  string           `json:"message"`
	Version  string           `json:"version"`
	Features *config.Features `json:"features,omitempty"`
}

type SaveBackupRequest struct {
	Uid       string                `form:"uid"`
	Version   string                `form:"version"`
	DeviceKey string                `form:"userkey"`
	Reason    string                `form:"reason"`
	Content   *multipart.FileHeader `form:"file"`
	Hash      string                `form:"sha256_hash"`
	Hint      string                `form:"hint"`
	ManualMax int                   `form:"manmax"`
	Date      *time.Time            `form:"date"`
}

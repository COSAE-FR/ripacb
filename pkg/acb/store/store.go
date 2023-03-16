package store

import (
	"errors"
	"fmt"
	"github.com/COSAE-FR/ripacb/pkg/acb/config"
	"github.com/COSAE-FR/ripacb/pkg/acb/entity"
)

var (
	ErrMaskedToClient    = errors.New("")
	ErrSaveDisabled      = fmt.Errorf("%wsave disabled", ErrMaskedToClient)
	ErrDeleteDisabled    = fmt.Errorf("%wdelete disabled", ErrMaskedToClient)
	ErrDeleteForbidden   = fmt.Errorf("%wdelete forbidden", ErrMaskedToClient)
	ErrNewDeviceDisabled = fmt.Errorf("%wnew devices disabled", ErrMaskedToClient)
)

type Store interface {
	GetRevisionsForDevice(device string, features config.Features) (entity.RevisionList, error)
	GetRevision(device string, revisionId string, features config.Features) (*entity.Revision, error)
	DeleteRevision(device string, revisionId string, features config.Features) error
	SetRevision(revision *entity.Revision, features config.Features) error
}


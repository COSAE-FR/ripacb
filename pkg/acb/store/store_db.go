package store

import (
	"errors"
	"fmt"
	"github.com/COSAE-FR/ripacb/pkg/acb"
	"github.com/COSAE-FR/ripacb/pkg/acb/config"
	"github.com/COSAE-FR/ripacb/pkg/acb/entity"
	"github.com/COSAE-FR/riputils/db"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

const DBStoreType acb.StoreType = "database"

type DBRevision struct {
	Revision   string `gorm:"primaryKey;autoIncrement:false"`
	Hash       string
	Content    string
	Reason     string
	Device     string `gorm:"primaryKey;autoIncrement:false"`
	Username   string
	Comment    string
	Date       time.Time
	FromPortal bool
}

func (DBRevision) TableName() string {
	return "acbd_revision"
}

func (r *DBRevision) ToRevision() *entity.Revision {
	return &entity.Revision{
		Revision:   r.Revision,
		Hash:       r.Hash,
		Content:    r.Content,
		Reason:     r.Reason,
		Device:     r.Device,
		Username:   r.Username,
		Comment:    r.Comment,
		Date:       r.Date,
		FromPortal: r.FromPortal,
	}
}

type DBStore struct {
	log *logrus.Entry
	db  *gorm.DB
}

func (D DBStore) GetRevisionsForDevice(device string, features config.Features) (entity.RevisionList, error) {
	var dbRevisions []DBRevision
	db := D.db
	if !features.AllowRestoreUser {
		db = db.Where("from_portal = ?", true)
	}
	db = db.Find(&dbRevisions, "device = ?", device)
	if db.Error != nil {
		return nil, db.Error
	}
	var list entity.RevisionList
	for _, r := range dbRevisions {
		list[r.Revision] = *r.ToRevision()
	}
	return list, nil
}

func (D DBStore) GetRevision(device string, revisionId string, features config.Features) (*entity.Revision, error) {
	var dbRevision DBRevision
	db := D.db
	if !features.AllowRestoreUser {
		db = db.Where("from_portal = ?", true)
	}
	err := db.First(&dbRevision, "revision = ? and device = ?", revisionId, device).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unknown revision")
	}
	return dbRevision.ToRevision(), nil
}

func (D DBStore) DeleteRevision(device string, revisionId string, features config.Features) error {
	// TODO: Implement
	return nil
}

func (D DBStore) SetRevision(revision *entity.Revision, features config.Features) error {
	if !features.AllowSave {
		return ErrSaveDisabled
	}
	if revision == nil {
		return fmt.Errorf("empty revision")
	}
	validate := validator.New()
	if err := validate.Struct(revision); err != nil {
		return err
	}
	revision.FromPortal = features.IsPortal
	var revisions []DBRevision
	if err := D.db.Order("date").Find(&revisions, "device = ?", revision.Device).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if !features.AllowNew {
				return ErrNewDeviceDisabled
			}
		} else {
			return err
		}
	}
	// TODO: limit saved revisions
	if len(revisions) > features.MaxBackups {
		
	}
	if !features.AllowNew {
		if err := D.db.First(&DBRevision{}, "device = ?", revision.Device).Error; err != nil {
			return ErrNewDeviceDisabled
		}
	}
	return D.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&DBRevision{
		Revision:   revision.Revision,
		Hash:       revision.Hash,
		Content:    revision.Content,
		Reason:     revision.Reason,
		Device:     revision.Device,
		Username:   revision.Username,
		Comment:    revision.Comment,
		Date:       revision.Date,
		FromPortal: revision.FromPortal,
	}).Error
}

func NewDBStore(dbConfig *db.Configuration, logger *logrus.Entry) (*DBStore, error) {
	logger = logger.WithFields(map[string]interface{}{
		"component": "db_store",
	})
	if dbConfig == nil {
		return nil, fmt.Errorf("empty DB configuration")
	}
	if err := dbConfig.Check(); err != nil {
		return nil, err
	}
	db, err := dbConfig.Open()
	if err != nil {
		return nil, err
	}
	_ = db.AutoMigrate(&DBRevision{})
	return &DBStore{db: db.Model(&DBRevision{}), log: logger}, nil
}

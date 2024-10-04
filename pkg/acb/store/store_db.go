package store

import (
	"errors"
	"fmt"
	"github.com/COSAE-FR/ripacb/pkg/acb/config"
	"github.com/COSAE-FR/ripacb/pkg/acb/entity"
	"github.com/COSAE-FR/riputils/db"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

const DBStoreType Type = "database"

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

func (*DBRevision) TableName() string {
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
	tx := D.db.Order("date desc")
	if !features.AllowRestoreUser {
		tx = tx.Where("from_portal = ?", true)
	}
	tx = tx.Find(&dbRevisions, "device = ?", device)
	if tx.Error != nil {
		return nil, tx.Error
	}
	list := make(entity.RevisionList, len(dbRevisions))
	for _, r := range dbRevisions {
		list[r.Revision] = *r.ToRevision()
	}
	return list, nil
}

func (D DBStore) GetRevision(device string, revisionId string, features config.Features) (*entity.Revision, error) {
	var dbRevision DBRevision
	database := D.db.Order("date desc")
	if !features.AllowRestoreUser {
		database = database.Where("from_portal = ?", true)
	}
	err := database.First(&dbRevision, "revision = ? and device = ?", revisionId, device).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("unknown revision")
		}
		return nil, err
	}
	return dbRevision.ToRevision(), nil
}

func (D DBStore) DeleteRevision(device string, revisionId string, features config.Features) error {
	if !features.AllowDelete {
		return ErrDeleteDisabled
	}
	return D.db.Transaction(func(tx *gorm.DB) error {
		return tx.Delete(&DBRevision{}, "revision = ? and device = ?", revisionId, device).Error
	})
}

func (D DBStore) SetRevision(revision *entity.Revision, features config.Features) error {
	if !features.AllowSave {
		return ErrSaveDisabled
	}
	if revision == nil {
		return fmt.Errorf("empty revision")
	}
	logger := D.log.WithFields(logrus.Fields{"method": "set_revision", "revision": revision.Revision})
	validate := validator.New()
	if err := validate.Struct(revision); err != nil {
		return err
	}
	revision.FromPortal = features.IsPortal
	return D.db.Transaction(func(tx *gorm.DB) error {
		var revisions []DBRevision
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Order("date desc").Find(&revisions, "device = ?", revision.Device).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if !features.AllowNew {
					return ErrNewDeviceDisabled
				}
			} else {
				return err
			}
		}
		if len(revisions) >= features.MaxBackups {
			toDelete := len(revisions) - features.MaxBackups + 1
			revisionsToDelete := make([]string, toDelete)
			for i := len(revisions) - 1; i >= features.MaxBackups-1; i-- {
				revisionsToDelete = append(revisionsToDelete, revisions[i].Revision)
			}
			logger.WithFields(logrus.Fields{
				"revisions":           len(revisions),
				"max":                 features.MaxBackups,
				"to_delete":           toDelete,
				"to_delete_revisions": revisionsToDelete,
			}).Tracef("Cleaning database")
			if err := tx.Delete(&DBRevision{}, "device = ? AND revision IN ?", revision.Device, revisionsToDelete).Error; err != nil {
				logger.Errorf("Cannot delete extraneous records: %s", err)
			}
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&DBRevision{
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
	})

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
	database, err := dbConfig.Open()
	if err != nil {
		return nil, err
	}
	_ = database.AutoMigrate(&DBRevision{})
	return &DBStore{db: database.Model(&DBRevision{}), log: logger}, nil
}

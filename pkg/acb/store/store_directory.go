package store

import (
	"encoding/json"
	"fmt"
	"github.com/COSAE-FR/ripacb/pkg/acb/config"
	"github.com/COSAE-FR/ripacb/pkg/acb/entity"
	"github.com/COSAE-FR/riputils/common"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

const DirectoryStoreType Type = "directory"

type DirectoryStore struct {
	log  *logrus.Entry
	path string
	lock sync.RWMutex
}

func (d *DirectoryStore) GetRevisionsForDevice(device string, features config.Features) (entity.RevisionList, error) {
	logger := d.log.WithFields(map[string]interface{}{
		"action": "get_revisions",
		"device": device,
	})
	devicePath := filepath.Join(d.path, device)
	d.lock.RLock()
	defer d.lock.RUnlock()
	if !common.IsDirectory(devicePath) {
		return nil, fmt.Errorf("unknown device")
	}
	revisions := make(entity.RevisionList)
	err := filepath.WalkDir(devicePath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".json" {
			logger.Tracef("Invalid file: %s", path)
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			logger.Errorf("Cannot read %s", err)
			return nil
		}
		var revision entity.Revision
		err = json.Unmarshal(content, &revision)
		if err != nil {
			logger.Errorf("Cannot parse %s", err)
			return nil
		}
		if revision.Revision != "" && revision.Content != "" {
			if !revision.FromPortal && !features.AllowRestoreUser {
				logger.Tracef("Not adding revision %s from %s: not portal revision", revision.Revision, path)
				return nil
			}
			logger.Tracef("Adding revision %s from %s", revision.Revision, path)
			revisions[revision.Revision] = revision
		} else {
			logger.Errorf("Invalid content for %s", path)
		}
		return nil
	})
	return revisions, err
}

func (d *DirectoryStore) GetRevision(device string, revisionId string, features config.Features) (*entity.Revision, error) {
	list, err := d.GetRevisionsForDevice(device, features)
	if err != nil {
		return nil, err
	}
	rev, found := list[revisionId]
	if !found {
		return nil, fmt.Errorf("unknown revision")
	}
	return &rev, nil
}

func (d *DirectoryStore) SetRevision(revision *entity.Revision, features config.Features) error {
	if revision == nil {
		return fmt.Errorf("empty revision")
	}
	validate := validator.New()
	if err := validate.Struct(revision); err != nil {
		return err
	}
	revision.FromPortal = features.IsPortal
	if !features.AllowSave {
		return ErrSaveDisabled
	}
	devicePath := filepath.Join(d.path, revision.Device)
	d.lock.Lock()
	defer d.lock.Unlock()
	if !common.IsDirectory(devicePath) {
		if features.AllowNew {
			if err := os.MkdirAll(devicePath, 0750); err != nil {
				return err
			}
		} else {
			return ErrNewDeviceDisabled
		}
	}
	existing, err := filepath.Glob(fmt.Sprintf("%s/*.json", devicePath))
	if err != nil {
		return err
	}
	for len(existing) >= features.MaxBackups && features.MaxBackups > 0 {
		if err := os.Remove(existing[0]); err != nil {
			return err
		}
		existing = existing[1:]
	}
	content, err := json.Marshal(revision)
	if err != nil {
		return err
	}
	backupPath := filepath.Join(devicePath, fmt.Sprintf("%s.json", revision.Revision))
	if err := os.WriteFile(backupPath, content, 0640); err != nil {
		return err
	}
	return nil
}

func (d *DirectoryStore) DeleteRevision(device string, revisionId string, features config.Features) error {
	if !features.AllowDelete {
		return ErrDeleteDisabled
	}
	list, err := d.GetRevisionsForDevice(device, features)
	if err != nil {
		return err
	}
	rev, found := list[revisionId]
	if !found {
		return fmt.Errorf("unknown revision")
	}
	if rev.FromPortal && !features.IsPortal {
		return ErrDeleteForbidden
	}
	devicePath := filepath.Join(d.path, rev.Device)
	d.lock.Lock()
	defer d.lock.Unlock()
	if common.FileExists(devicePath) {
		return os.Remove(devicePath)
	}
	return nil
}

func NewDirectoryStore(path string, logger *logrus.Entry) (*DirectoryStore, error) {
	logger = logger.WithFields(map[string]interface{}{
		"component":  "directory_store",
		"store_path": path,
	})
	if !common.IsDirectory(path) {
		err := os.MkdirAll(path, 0750)
		if err != nil {
			return nil, err
		}
	}
	return &DirectoryStore{log: logger, path: path}, nil
}

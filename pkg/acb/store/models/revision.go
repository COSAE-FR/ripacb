package models

import (
	"time"
)

type Revision struct {
	Id         int64  `json:"id" gorm:"primary_key"`
	Revision   string `validate:"required,min=6" gorm:"uniqueIndex:idx_revision;not null"`
	Hash       string `validate:"hexadecimal,required,len=64"`
	Content    string `validate:"required"`
	Reason     string `validate:"required"`
	Device     string `validate:"hexadecimal,required,len=64" gorm:"uniqueIndex:idx_revision;not null"`
	Username   string
	Comment    string
	Date       time.Time
	FromPortal bool
}

func (Revision) TableName() string {
	return "acbd_revision"
}

package models

import (
	"time"

	"gorm.io/gorm"
)

type CommonFields struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

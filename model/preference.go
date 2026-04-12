package model

import (
	"time"

	"gorm.io/gorm"
)

type UserPreference struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"index;not null"`
	AIType uint `gorm:"not null"`
	Score  int  `gorm:"not null;default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

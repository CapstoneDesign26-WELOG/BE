package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID             uint    `gorm:"primaryKey"`
	Email          string  `gorm:"type:varchar(100);unique;not null"`
	HashedPassword *string `gorm:"type:varchar(255)"`

	Nickname   string `gorm:"type:varchar(50);not null"`
	Provider   string `gorm:"type:varchar(20);not null;default:'google'"`
	ProviderID string `gorm:"type:varchar(255);uniqueIndex;not null"`
	TokenCount int    `gorm:"default:0"`

	Preferences []UserPreference `gorm:"foreignKey:UserID"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

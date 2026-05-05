package model

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `gorm:"index;not null"`
	Title       string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text;not null"`
	Type        uint   `gorm:"not null"`
	Count       uint   `gorm:"not null"`

	Comments []Comment `gorm:"foreignKey:PostID"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

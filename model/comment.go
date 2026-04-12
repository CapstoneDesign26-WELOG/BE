package model

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID       uint  `gorm:"primaryKey"`
	PostID   uint  `gorm:"index;not null"`
	UserID   uint  `gorm:"index;not null"`
	ParentID *uint `gorm:"index"`

	Description string `gorm:"type:text;not null"`
	IsAI        bool   `gorm:"default:false;not null"`
	AIType      *uint

	User   User     `gorm:"foreignKey:UserID"`
	Post   Post     `gorm:"foreignKey:PostID"`
	Parent *Comment `gorm:"foreignKey:ParentID"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

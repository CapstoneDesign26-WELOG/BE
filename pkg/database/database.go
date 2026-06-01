package database

import (
	"log"
	"welog/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ DB 연결 실패: %v", err)
	}

	log.Println("🔄 DB 마이그레이션 진행 중...")
	err = db.AutoMigrate(
		&model.User{},
		&model.UserPreference{},
		&model.Post{},
		&model.Comment{},
		&model.CommentLike{},
	)
	if err != nil {
		log.Fatalf("❌ DB 마이그레이션 실패: %v", err)
	}

	log.Println("✅ DB 연결 및 마이그레이션 성공!")
	return db
}

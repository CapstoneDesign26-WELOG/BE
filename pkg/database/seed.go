package database

import (
	"log"
	"welog/internal/model"

	"gorm.io/gorm"
)

func SeedSystemUser(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Where("id = ?", 1).Count(&count)
	if count == 0 {
		log.Println("시스템 유저(ID: 1)가 존재하지 않아 생성을 시작합니다...")
		systemUser := model.User{
			ID:         1,
			Email:      "system@welog.ai",
			Nickname:   "AI Assisatant",
			Provider:   "system",
			ProviderID: "system-ai-account",
			Role:       "SYSTEM",
			TokenCount: 0,
		}
		if err := db.Create(&systemUser).Error; err != nil {
			log.Printf("시스템 유저 생성 실패: %v", err)
		} else {
			log.Println("시스템 유저(ID: 1)가 성공적으로 생성되었습니다.")
		}
	}
}

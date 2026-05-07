package user

import (
	"errors"
	"welog/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(userID uint) (*model.User, error) {
	var user model.User
	err := r.db.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) UpdateUserToken(userID uint, newCount int) error {
	return r.db.Model(&model.User{}).Where("user_id = ?", userID).Update("token_count", newCount).Error
}

func (r *UserRepository) UpdateTokenBelowThreshold(threshold, targetCount uint) error {
	return r.db.Model(&model.User{}).Where("token_count < ?", threshold).Update("token_count", targetCount).Error
}

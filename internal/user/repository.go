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
	err := r.db.Where("id = ?", userID).First(&user).Error
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

func (r *UserRepository) ConsumeToken(userID, amount uint) error {
	result := r.db.Model(&model.User{}).
		Where("id = ? AND token_count >= ?", userID, amount).
		UpdateColumn("token_count", gorm.Expr("token_count - ?", amount))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("토큰이 부족합니다")
	}
	return nil
}

func (r *UserRepository) UpdateTokenBelowThreshold(threshold, targetCount uint) error {
	return r.db.Model(&model.User{}).Where("token_count < ?", threshold).Update("token_count", targetCount).Error
}

func (r *UserRepository) Save(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) RefundToken(userID, amount uint) error {
	return r.db.Model(&model.User{}).
		Where("id = ? ", userID).
		UpdateColumn("token_count", gorm.Expr("token_count + ?", amount)).Error
}

func (r *UserRepository) GetUserPreferences(userID uint) (map[uint]int, error) {
	var results []struct {
		AIType uint
		Score  int
	}
	err := r.db.Model(model.UserPreference{}).
		Select("ai_type, score").
		Where("user_id = ?", userID).
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	prefs := make(map[uint]int)
	for _, res := range results {
		prefs[res.AIType] = res.Score
	}
	return prefs, nil
}

package user

import "welog/internal/model"

type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(userID uint) (*model.User, error) {
	return s.repo.FindByID(userID)
}

func (s *UserService) ConsumeToken(userID uint, newCount int) error {
	return s.repo.UpdateUserToken(userID, newCount)
}

func (s *UserService) RefillDailyTokens() error {
	const threshold uint = 3
	const targetCount uint = 3
	return s.repo.UpdateTokenBelowThreshold(threshold, targetCount)
}

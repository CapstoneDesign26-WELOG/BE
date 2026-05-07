package user

type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) RefillDailyTokens() error {
	const threshold uint = 3
	const targetCount uint = 3
	return s.repo.UpdateTokenBelowThreshold(threshold, targetCount)
}

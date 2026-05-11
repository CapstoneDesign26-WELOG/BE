package auth

import (
	"context"
	"errors"
	"welog/internal/model"

	"google.golang.org/api/idtoken"
)

type UserRepository interface {
	FindByEmail(email string) (*model.User, error)
	Create(user *model.User) error
}

type AuthService struct {
	UserRepo       UserRepository
	jwtSecret      []byte
	googleClientID string
}

func NewAuthService(userRepo UserRepository, jwtSecret string, googleClientID string) *AuthService {
	return &AuthService{
		UserRepo:       userRepo,
		jwtSecret:      []byte(jwtSecret),
		googleClientID: googleClientID,
	}
}

func (s *AuthService) ProcessGoogleLogin(ctx context.Context, googleToken string) (string, *model.User, error) {
	payload, err := idtoken.Validate(ctx, googleToken, s.googleClientID)
	if err != nil {
		return "", nil, errors.New("invalid google token")
	}

	email := payload.Claims["email"].(string)
	providerID := payload.Subject

	userObj, err := s.UserRepo.FindByEmail(email)
	if err != nil {
		return "", nil, err
	}

	if userObj == nil {
		newUser := &model.User{
			Email:      email,
			Provider:   "google",
			ProviderID: providerID,
			Nickname:   "User_" + providerID[:6],
			Role:       "USER",
		}
		if err := s.UserRepo.Create(newUser); err != nil {
			return "", nil, err
		}
		userObj = newUser
	}

	accessToken, err := GenerateToken(userObj.ID, userObj.Email, userObj.Role, s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return accessToken, userObj, nil
}

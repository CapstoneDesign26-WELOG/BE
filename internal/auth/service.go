package auth

import (
	"context"
	"errors"
	"welog/internal/model"
	"welog/internal/user"

	"cloud.google.com/go/auth/credentials/idtoken"
)

type AuthService struct {
	UserRepo *user.UserRepository
}

func NewAuthService(userRepo *user.UserRepository) *AuthService {
	return &AuthService{UserRepo: userRepo}
}

func (s *AuthService) ProcessGoogleLogin(ctx context.Context, googleToken string) (string, *model.User, error) {
	googleClientID := ""

	payload, err := idtoken.Validate(ctx, googleToken, googleClientID)
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
			ProviderID: userObj.ProviderID,
			Nickname:   "User_" + providerID[:6],
			Role:       "USER",
		}
		if err := s.UserRepo.Create(newUser); err != nil {
			return "", nil, err
		}
		userObj = newUser
	}

	accessToekn, err := GenerateToken(userObj.ID, userObj.Email, userObj.Role)
	if err != nil {
		return "", nil, err
	}

	return accessToekn, userObj, nil
}

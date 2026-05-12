package auth

import (
	"context"
	"errors"
	"os"
	"strings"
	"welog/internal/model"

	"google.golang.org/api/idtoken"
)

type UserRepository interface {
	FindByEmail(email string) (*model.User, error)
	Create(user *model.User) error
	Save(user *model.User) error
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

func isAdmin(email string) bool {
	adminStr := os.Getenv("ADMIN_EMAILS")
	if adminStr == "" {
		return false
	}
	admins := strings.Split(adminStr, ",")
	for _, adminEmail := range admins {
		if strings.TrimSpace(adminEmail) == email {
			return true
		}
	}
	return false
}

func (s *AuthService) ProcessGoogleLogin(ctx context.Context, googleToken string) (string, *model.User, error) {
	payload, err := idtoken.Validate(ctx, googleToken, s.googleClientID)
	if err != nil {
		return "", nil, errors.New("invalid google token")
	}

	email := payload.Claims["email"].(string)
	providerID := payload.Subject

	role := "USER"
	tokenCount := 10
	if isAdmin(email) {
		role = "ADMIN"
		tokenCount = 99999
	}

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
			Role:       role,
			TokenCount: tokenCount,
		}
		if err := s.UserRepo.Create(newUser); err != nil {
			return "", nil, err
		}
		userObj = newUser
	} else {
		if role == "ADMIN" && userObj.Role != "ADMIN" {
			userObj.Role = "ADMIN"
			userObj.TokenCount = 99999
			s.UserRepo.Save(userObj)
		}
	}

	accessToken, err := GenerateToken(userObj.ID, userObj.Email, userObj.Role, s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return accessToken, userObj, nil
}

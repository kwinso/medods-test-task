package services

import (
	"github.com/kwinso/medods-test-task/internal/db/repositories"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthService interface {
	AuthorizeByGUID(guid string) (*TokenPair, error)
}

type authService struct {
	repo repositories.AuthRepository
}

func NewAuthService(repo repositories.AuthRepository) AuthService {
	return &authService{
		repo: repo,
	}
}

func (s *authService) AuthorizeByGUID(guid string) (*TokenPair, error) {
	return &TokenPair{}, nil
}

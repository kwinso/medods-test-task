package services

import (
	"context"
	"net/netip"
	"time"

	"github.com/kwinso/medods-test-task/internal/db"
	"github.com/kwinso/medods-test-task/internal/db/repositories"
	"github.com/kwinso/medods-test-task/internal/tokens"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthService interface {
	AuthorizeByGUID(ctx context.Context, guid, userAgent string, ip netip.Addr) (*TokenPair, error)
}

type authService struct {
	repo repositories.AuthRepository
	key  string
}

func NewAuthService(repo repositories.AuthRepository, key string) AuthService {
	return &authService{
		repo: repo,
		key:  key,
	}
}

func (s *authService) AuthorizeByGUID(ctx context.Context, guid, userAgent string, ip netip.Addr) (*TokenPair, error) {
	refreshToken, err := tokens.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshTokenHash, err := tokens.HashRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	auth, err := s.repo.CreateAuth(ctx, db.CreateAuthParams{
		Guid:             guid,
		RefreshTokenHash: refreshTokenHash,
		IpAddress:        ip,
		UserAgent:        userAgent,
		RefreshedAt:      time.Now(),
	})

	if err != nil {
		return nil, err
	}

	accessToken, err := tokens.GenerateAccessToken(auth.Guid, (int)(auth.ID), s.key)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

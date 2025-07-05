package services

import (
	"context"
	"database/sql"
	"errors"
	"net/netip"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kwinso/medods-test-task/internal/db"
	"github.com/kwinso/medods-test-task/internal/db/repositories"
	"github.com/kwinso/medods-test-task/internal/tokens"
)

var ErrAuthExpired = errors.New("auth expired")

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthService interface {
	AuthorizeByGUID(ctx context.Context, guid, userAgent string, ip netip.Addr) (*TokenPair, error)
	GetAuthByToken(ctx context.Context, token string) (*db.Auth, error)
	DeleteAuthById(ctx context.Context, authId int32) error
}

type authService struct {
	repo     repositories.AuthRepository
	key      string
	tokenTTL time.Duration
	authTTL  time.Duration
}

func NewAuthService(repo repositories.AuthRepository, key string, tokenTTL, authTTL time.Duration) AuthService {
	return &authService{
		repo:     repo,
		key:      key,
		tokenTTL: tokenTTL,
		authTTL:  authTTL,
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

	accessToken, err := tokens.GenerateAccessToken(auth.Guid, (int)(auth.ID), s.key, s.tokenTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) GetAuthByToken(ctx context.Context, token string) (*db.Auth, error) {
	claims, err := tokens.ParseAccessToken(token, s.key)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrAuthExpired
		}
		return nil, err
	}

	auth, err := s.repo.GetAuthById(ctx, int32(claims.AuthId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAuthExpired
		}
		return nil, err
	}

	// If refreshed way to long ago, this auth is no longer valid
	if time.Now().After(auth.RefreshedAt.Add(s.authTTL)) {
		return nil, ErrAuthExpired
	}

	return &auth, nil
}

func (s *authService) DeleteAuthById(ctx context.Context, authId int32) error {
	return s.repo.DeleteAuthById(ctx, authId)
}

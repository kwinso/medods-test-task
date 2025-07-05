package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kwinso/medods-test-task/internal/db"
	"github.com/kwinso/medods-test-task/internal/db/repositories"
	"github.com/kwinso/medods-test-task/internal/tokens"
)

var (
	ErrAuthExpired        = errors.New("auth expired")
	ErrUserAgentMismatch  = errors.New("user agent mismatch")
	ErrInvalidTokenFormat = errors.New("invalid token format")
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthService interface {
	AuthorizeByGUID(ctx context.Context, guid, userAgent string, ip netip.Addr) (*TokenPair, error)
	GetAuthByAccessToken(ctx context.Context, token string) (*db.Auth, error)
	// RefreshAuth refreshes the access token for the user.
	//
	// Returns:
	// 	- ErrAuthExpired if the refresh token is expired
	// 	- ErrUserAgentMismatch if the user agent does not match. Mismatched user agent causes auth to be dropped
	RefreshAuth(ctx context.Context, refreshToken, userAgent string, ip netip.Addr) (*TokenPair, error)
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
	nextId, err := s.repo.GetNextAuthId(ctx)
	if err != nil {
		return nil, err
	}

	refreshToken, err := tokens.GenerateRefreshToken(int(nextId), s.key)
	if err != nil {
		return nil, err
	}
	fmt.Println(refreshToken)

	refreshTokenHash, err := tokens.HashRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	auth, err := s.repo.CreateAuth(ctx, db.CreateAuthParams{
		ID:               int32(nextId),
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

func (s *authService) GetAuthByAccessToken(ctx context.Context, token string) (*db.Auth, error) {
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

func (s *authService) RefreshAuth(ctx context.Context, encodedRefreshToken, userAgent string, ip netip.Addr) (*TokenPair, error) {
	authId, err := tokens.ParseEncodedRefreshToken(encodedRefreshToken, s.key)
	if err != nil {
		if errors.Is(err, tokens.ErrInvalidTokenFormat) {
			return nil, ErrInvalidTokenFormat
		}
		return nil, err
	}

	auth, err := s.repo.GetAuthById(ctx, int32(authId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAuthExpired
		}
		return nil, err
	}

	if time.Now().After(auth.RefreshedAt.Add(s.authTTL)) {
		return nil, ErrAuthExpired
	}

	if auth.UserAgent != userAgent {
		_ = s.DeleteAuthById(ctx, auth.ID)
		return nil, ErrUserAgentMismatch
	}

	if auth.IpAddress != ip {
		// TODO: Send webhook
	}

	refreshToken, err := tokens.GenerateRefreshToken(int(auth.ID), s.key)
	if err != nil {
		return nil, err
	}
	refreshTokenHash, err := tokens.HashRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	err = s.repo.UpdateAuthRefreshToken(ctx, auth.ID, refreshTokenHash)
	if err != nil {
		return nil, err
	}

	accessToken, err := tokens.GenerateAccessToken(auth.Guid, (int)(auth.ID), s.key, s.tokenTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: encodedRefreshToken,
	}, nil
}

func (s *authService) DeleteAuthById(ctx context.Context, authId int32) error {
	return s.repo.DeleteAuthById(ctx, authId)
}

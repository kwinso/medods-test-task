package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log"
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
	DeleteAuthById(ctx context.Context, authId uuid.UUID) error
}

type authService struct {
	repo          repositories.AuthRepository
	key           string
	tokenTTL      time.Duration
	authTTL       time.Duration
	logger        *log.Logger
	reportService ReportService
}

func NewAuthService(repo repositories.AuthRepository, reportService ReportService, logger *log.Logger, key string, tokenTTL, authTTL time.Duration) AuthService {
	return &authService{
		repo:          repo,
		key:           key,
		tokenTTL:      tokenTTL,
		authTTL:       authTTL,
		logger:        logger,
		reportService: reportService,
	}
}

func (s *authService) AuthorizeByGUID(ctx context.Context, guid, userAgent string, ip netip.Addr) (*TokenPair, error) {
	recordId := uuid.New()
	refreshToken, err := tokens.GenerateRefreshToken(recordId)
	if err != nil {
		return nil, err
	}
	fmt.Println(refreshToken)

	refreshTokenHash, err := tokens.HashRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	auth, err := s.repo.CreateAuth(ctx, db.CreateAuthParams{
		ID:               recordId,
		Guid:             guid,
		RefreshTokenHash: refreshTokenHash,
		IpAddress:        ip,
		UserAgent:        userAgent,
		RefreshedAt:      time.Now(),
	})

	if err != nil {
		return nil, err
	}

	accessToken, err := tokens.GenerateAccessToken(auth.Guid, auth.ID, s.key, s.tokenTTL)
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

	auth, err := s.repo.GetAuthById(ctx, claims.AuthId)
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

func (s *authService) RefreshAuth(ctx context.Context, refreshToken, userAgent string, ip netip.Addr) (*TokenPair, error) {
	authId, err := tokens.ParseEncodedRefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, tokens.ErrInvalidTokenFormat) {
			return nil, ErrInvalidTokenFormat
		}
		return nil, err
	}

	auth, err := s.repo.GetAuthById(ctx, *authId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAuthExpired
		}
		return nil, err
	}

	valid := tokens.VerifyRefreshToken(refreshToken, auth.RefreshTokenHash)
	if !valid {
		return nil, ErrAuthExpired
	}

	if time.Now().After(auth.RefreshedAt.Add(s.authTTL)) {
		return nil, ErrAuthExpired
	}

	if auth.UserAgent != userAgent {
		s.logger.Printf("User agent mismatch (was %q, got %q) for user %v. Dropping authorization", auth.UserAgent, userAgent, auth.Guid)
		_ = s.DeleteAuthById(ctx, auth.ID)
		return nil, ErrUserAgentMismatch
	}

	if auth.IpAddress.Compare(ip) != 0 {
		s.logger.Printf("Auth for %v IP changed from %q to %q. Sending report to webhook.", auth.Guid, auth.IpAddress, ip)
		err := s.reportService.SendIPChangeReport(auth, ip)
		if err != nil {
			s.logger.Printf("Failed to deliver webhook change report for %v: %v", auth.Guid, err)
		}
	}

	newRefreshToken, err := tokens.GenerateRefreshToken(auth.ID)
	if err != nil {
		return nil, err
	}
	refreshTokenHash, err := tokens.HashRefreshToken(newRefreshToken)
	if err != nil {
		return nil, err
	}

	err = s.repo.UpdateAuthRefreshToken(ctx, auth.ID, refreshTokenHash)
	if err != nil {
		return nil, err
	}

	accessToken, err := tokens.GenerateAccessToken(auth.Guid, auth.ID, s.key, s.tokenTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) DeleteAuthById(ctx context.Context, authId uuid.UUID) error {
	return s.repo.DeleteAuthById(ctx, authId)
}

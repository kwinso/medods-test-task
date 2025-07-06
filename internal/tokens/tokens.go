package tokens

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/google/uuid"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidTokenFormat = errors.New("invalid token format")
)

type TokenClaims struct {
	jwt.RegisteredClaims
	Guid   string    `json:"guid"`
	AuthId uuid.UUID `json:"auth_id"`
}

func GenerateAccessToken(guid string, authId uuid.UUID, key string, ttl time.Duration) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
		Guid:   guid,
		AuthId: authId,
	})

	return t.SignedString([]byte(key))
}

func ParseAccessToken(tokenString, key string) (*TokenClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counterpart to verify
		return []byte(key), nil
	})
	if err != nil {
		return nil, err
	}

	return token.Claims.(*TokenClaims), nil
}

// GenerateRefreshToken generates a new refresh token which is a random 32-bit base64 string
func GenerateRefreshToken(authId uuid.UUID) (string, error) {
	payload := authId.String()
	randomPart := make([]byte, 16)
	_, err := rand.Read(randomPart)
	if err != nil {
		return "", err
	}
	randomPartHex := hex.EncodeToString(randomPart)
	token := "rt." + payload + "." + randomPartHex
	return token, nil
}

func EncodeRefreshTokenToBase64(token string) string {
	return base64.StdEncoding.EncodeToString([]byte(token))
}

func ParseEncodedRefreshToken(token string) (*uuid.UUID, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidTokenFormat
	}
	strUUID := parts[1]
	parsedUUID, err := uuid.Parse(strUUID)
	if err != nil {
		return nil, ErrInvalidTokenFormat
	}
	return &parsedUUID, nil
}

// HashRefreshToken hashes a refresh token using bcrypt for storing in the database
func HashRefreshToken(refreshToken string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(refreshToken), 5)
	return string(bytes), err
}

// VerifyRefreshToken verifies a refresh token using bcrypt
func VerifyRefreshToken(refreshToken, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(refreshToken))
	return err == nil
}

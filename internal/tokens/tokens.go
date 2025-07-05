package tokens

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type TokenClaims struct {
	jwt.RegisteredClaims
	Guid   string `json:"guid"`
	AuthId int    `json:"auth_id"`
}

func GenerateAccessToken(guid string, authId int, key string, ttl time.Duration) (string, error) {
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
		// we also only use its public counter part to verify
		return []byte(key), nil
	})
	if err != nil {
		return nil, err
	}

	return token.Claims.(*TokenClaims), nil
}

// GenerateRefreshToken generates a new refresh token which is a random 32-bit base64 string
func GenerateRefreshToken() (string, error) {
	var bytes [32]byte

	_, err := rand.Read(bytes[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bytes[:]), nil
}

// HashRefreshToken hashes a refresh token using bcrypt for storing in the database
func HashRefreshToken(refreshToken string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(refreshToken), 14)
	return string(bytes), err
}

// VerifyRefreshToken verifies a refresh token using bcrypt
func VerifyRefreshToken(refreshToken, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(refreshToken))
	return err == nil
}

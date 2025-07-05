package tokens

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidTokenFormat = errors.New("invalid token format")
	ErrInvalidSignature   = errors.New("invalid signature")
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
func GenerateRefreshToken(authId int, key string) (string, error) {
	payload := strconv.Itoa(authId)
	signature := sha256.Sum256([]byte(payload + key))
	hexSignature := hex.EncodeToString(signature[:])
	token := payload + "." + hexSignature
	return token, nil
}

func EncodeRefreshTokenToBase64(token string) string {
	return base64.StdEncoding.EncodeToString([]byte(token))
}

func ParseEncodedRefreshToken(encodedToken string, key string) (int, error) {
	tokenBytes, err := base64.StdEncoding.DecodeString(encodedToken)
	if err != nil {
		return 0, err
	}
	parts := strings.Split(string(tokenBytes), ".")
	if len(parts) != 2 {
		return 0, ErrInvalidTokenFormat
	}
	payload := parts[0]
	signature := parts[1]
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return 0, ErrInvalidTokenFormat
	}

	testSignature := sha256.Sum256([]byte(payload + key))

	if !hmac.Equal(signatureBytes, testSignature[:]) {
		return 0, ErrInvalidSignature
	}

	authId, err := strconv.Atoi(payload)
	if err != nil {
		return 0, err
	}
	return authId, nil
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

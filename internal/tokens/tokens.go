package tokens

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func GenerateAccessToken(guid string, authId int, key string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"guid":    guid,
		"auth_id": authId,
	})

	return t.SignedString([]byte(key))
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

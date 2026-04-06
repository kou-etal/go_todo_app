package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

type AccessTokenGenerator interface {
	GenerateAccessToken(userID string) (string, error)
}

type AccessTokenParser interface {
	ParseAccessToken(token string) (string, error)
}

type JWTManager struct {
	secret    []byte
	accessTTL time.Duration
}

func NewJWTManager(secret string, accessTTLSec int) *JWTManager {
	return &JWTManager{
		secret:    []byte(secret),
		accessTTL: time.Duration(accessTTLSec) * time.Second,
	}
}

func (m *JWTManager) GenerateAccessToken(userID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        uuid.NewString(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) ParseAccessToken(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	return claims.Subject, nil
}

package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	jwtKey []byte
}

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{
		jwtKey: []byte(secret),
	}
}

func (a *AuthService) GenerateToken(email string) (string, error) {

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(a.jwtKey)
}

func (a *AuthService) ParseToken(tokenString string) (string, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return a.jwtKey, nil
		},
	)

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)

	if !ok {
		return "", errors.New("invalid claims")
	}

	return claims.Email, nil
}

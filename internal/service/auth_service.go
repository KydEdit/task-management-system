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
	Email  string `json:"email"`
	UserID int    `json:"user_id"`
	jwt.RegisteredClaims
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{
		jwtKey: []byte(secret),
	}
}

func (a *AuthService) GenerateToken(email string, userID int) (string, error) {

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		Email:  email,
		UserID: userID,
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

func (a *AuthService) ParseToken(tokenString string) (string, int, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return a.jwtKey, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
	)

	if err != nil {
		return "", 0, err
	}

	if !token.Valid {
		return "", 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)

	if !ok {
		return "", 0, errors.New("invalid claims")
	}

	return claims.Email, claims.UserID, nil
}

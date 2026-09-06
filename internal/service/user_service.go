package service

import (
	"errors"
	"fmt"
	"task-manager-api/internal/models"
	"task-manager-api/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) Register(user models.RegisterRequest) (models.UserResponse, error) {
	var resp models.UserResponse

	if len(user.Password) < 8 || len(user.Password) > 72 {
		return models.UserResponse{}, models.ErrInvalidPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.UserResponse{}, err
	}
	hashedPassword := string(hashed)

	id, err := s.repo.RegisterUser(user.Email, hashedPassword)
	if err != nil {
		return models.UserResponse{}, err
	}

	resp.ID = id
	resp.Email = user.Email
	return resp, nil
}

func (s *UserService) Login(email, password string) (models.UserResponse, error) {
	var resp models.UserResponse

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return models.UserResponse{}, models.ErrInvalidCredentials
		}

		return models.UserResponse{}, fmt.Errorf("get user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return models.UserResponse{}, models.ErrInvalidCredentials
	}

	resp.ID = user.ID
	resp.Email = user.Email
	return resp, nil
}

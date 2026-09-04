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

func (s *UserService) Register(user models.User) (models.User, error) {

	if len(user.Password) < 8 {
		return models.User{}, errors.New("password must be at least 8 characters")
	}
	if len(user.Password) > 72 {
		return models.User{}, errors.New("password too long (max 72 bytes)")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	hashedPassword := string(hashed)

	id, err := s.repo.RegisterUser(user.Email, hashedPassword)
	if err != nil {
		return models.User{}, err
	}

	user.ID = id
	return user, nil
}

func (s *UserService) Login(email, password string) (models.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return models.User{}, models.ErrInvalidCredentials
		}

		return models.User{}, fmt.Errorf("get user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)
	if err != nil {
		return models.User{}, models.ErrInvalidCredentials
	}

	return user, nil
}

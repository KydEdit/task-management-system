package service

import "task-manager-api/internal/models"

type UserRepository interface {
	RegisterUser(email, password string, companyID int) (int, error)
	GetByEmail(email string) (models.User, error)
}

type CompanyRepository interface {
	EnsureExists(companyID int) error
}

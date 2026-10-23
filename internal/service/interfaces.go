package service

import (
	"context"
	"task-manager-api/internal/models"
)

type UserRepository interface {
	RegisterUser(ctx context.Context, email, password string, companyID int) (int, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
}

type CompanyRepository interface {
	EnsureExists(ctx context.Context, companyID int) error
}

type TaskRepository interface {
	CreateTask(ctx context.Context, title string, description string, email string, completed bool) (int, error)
	DeleteByID(ctx context.Context, email string, id int) error
	EditByID(ctx context.Context, email string, id int, task models.UserTasks) error
	GetAllByUser(ctx context.Context, email string) ([]models.UserTasks, error)
	GetByID(ctx context.Context, email string, id int) (models.UserTasks, error)
}

package handler

import (
	"context"
	"task-manager-api/internal/models"
)

type UserService interface {
	Login(ctx context.Context, email string, password string) (models.UserResponse, error)
	Me(ctx context.Context, email string) (models.UserResponse, error)
	Register(ctx context.Context, user models.RegisterRequest) (models.UserResponse, error)
}

type AuthService interface {
	GenerateToken(email string, userID int) (string, error)
	ParseToken(tokenString string) (string, int, error)
}

type TaskService interface {
	Create(ctx context.Context, task models.UserTasks) (models.UserTasks, error)
	Delete(ctx context.Context, userID int, id int) error
	Search(ctx context.Context, userID int) ([]models.UserTasks, error)
	SearchTask(ctx context.Context, userID int, id int) (models.UserTasks, error)
	Update(ctx context.Context, userID int, id int, task models.UserTasks) error
}

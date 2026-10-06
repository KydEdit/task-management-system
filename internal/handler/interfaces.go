package handler

import "task-manager-api/internal/models"

type UserService interface {
	Login(email string, password string) (models.UserResponse, error)
	Me(email string) (models.UserResponse, error)
	Register(user models.RegisterRequest) (models.UserResponse, error)
}

type AuthService interface {
	GenerateToken(email string) (string, error)
	ParseToken(tokenString string) (string, error)
}

type TaskService interface {
	Create(task models.UserTasks) (models.UserTasks, error)
	Delete(email string, id int) error
	Search(email string) ([]models.UserTasks, error)
	SearchTask(email string, id int) (models.UserTasks, error)
	Update(email string, id int, task models.UserTasks) error
}

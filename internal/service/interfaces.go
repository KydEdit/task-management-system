package service

import "task-manager-api/internal/models"

type UserRepository interface {
	RegisterUser(email, password string, companyID int) (int, error)
	GetByEmail(email string) (models.User, error)
}

type CompanyRepository interface {
	EnsureExists(companyID int) error
}

type TaskRepository interface {
	CreateTask(title string, description string, email string, completed bool) (int, error)
	DeleteByID(email string, id int) error
	EditByID(email string, id int, task models.UserTasks) error
	GetAllByUser(email string) ([]models.UserTasks, error)
	GetByID(email string, id int) (models.UserTasks, error)
}

package models

import "time"

type User struct {
	ID           int
	Email        string
	PasswordHash string
	CompanyID    int
}

type Company struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	CompanyID int    `json:"company_id"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	CompanyID int    `json:"company_id"`
}

type UserTasks struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	UserEmail     string `json:"-"`
	TaskCompleted bool   `json:"completed"`
}

package models

type User struct {
	ID           int
	Email        string
	PasswordHash string
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type UserTasks struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	UserEmail     string `json:"-"`
	TaskCompleted bool   `json:"completed"`
}

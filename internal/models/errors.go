package models

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrTaskNotFound       = errors.New("task not found or access denied")
)

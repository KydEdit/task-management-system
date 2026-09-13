package models

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrTaskNotFound       = errors.New("task not found or access denied")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidCompanyID   = errors.New("invalid company")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrCompanyNotFound    = errors.New("company not found")
)

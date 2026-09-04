package repository

import (
	"context"
	"errors"
	"fmt"
	"task-manager-api/internal/models"

	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	conn *pgx.Conn
}

func NewUserRepository(conn *pgx.Conn) *UserRepository {
	return &UserRepository{
		conn: conn,
	}
}

func (r *UserRepository) RegisterUser(email, password string) (int, error) {
	var id int

	err := r.conn.QueryRow(
		context.Background(),
		`
		INSERT INTO users (user_email, password)
		VALUES ($1, $2)
		RETURNING id
		`,
		email,
		password,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *UserRepository) GetByEmail(email string) (models.User, error) {
	var user models.User

	err := r.conn.QueryRow(
		context.Background(),
		`
		SELECT id, user_email, password
		FROM users
		WHERE user_email = $1
		`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

package repository

import (
	"context"
	"errors"
	"fmt"
	"task-manager-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	var pgErr *pgconn.PgError

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
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
				return 0, models.ErrUserAlreadyExists
			}
		}

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
		&user.PasswordHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

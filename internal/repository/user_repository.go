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

func (r *UserRepository) RegisterUser(ctx context.Context, email, password string, companyID int) (int, error) {
	var id int
	var pgErr *pgconn.PgError

	err := r.conn.QueryRow(
		ctx,
		`
		INSERT INTO users (email, password, company_id)
		VALUES ($1, $2, $3)
		RETURNING id
		`,
		email,
		password,
		companyID,
	).Scan(&id)

	if err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, models.ErrUserAlreadyExists
			}
		}

		return 0, err
	}

	return id, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User

	err := r.conn.QueryRow(
		ctx,
		`
		SELECT id, email, password, company_id
		FROM users
		WHERE email = $1
		`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CompanyID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

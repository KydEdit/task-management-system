package repository

import (
	"context"
	"errors"
	"fmt"
	"task-manager-api/internal/models"

	"github.com/jackc/pgx/v5"
)

type CompanyRepository struct {
	conn *pgx.Conn
}

func NewCompanyRepository(conn *pgx.Conn) *CompanyRepository {
	return &CompanyRepository{
		conn: conn,
	}
}

func (r *CompanyRepository) EnsureExists(companyID int) error {
	var dummy int

	err := r.conn.QueryRow(
		context.Background(),
		`
		SELECT id
		FROM companies
		WHERE id = $1
		`,
		companyID,
	).Scan(&dummy)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ErrCompanyNotFound
		}
		return fmt.Errorf("ensure company exists: %w", err)
	}

	return nil
}

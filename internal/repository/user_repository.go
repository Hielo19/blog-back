package repository

import (
	"context"

	"gin-demo/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
}

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
INSERT INTO users (
    username, email, password_hash, display_name, avatar_url, bio, role, status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, last_login_at, created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.DisplayName,
		user.AvatarURL,
		user.Bio,
		user.Role,
		user.Status,
	).Scan(
		&user.ID,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
}

package repository

import (
	"context"
	"errors"
	"fmt"

	"agentsight/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepo(pool *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{pool: pool}
}

func (r *PostgresUserRepo) GetByID(ctx context.Context, id int) (*models.User, error) {
	query := `SELECT id, github_id, username, avatar_url, email, created_at FROM users WHERE id = $1`
	var u models.User
	err := r.pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.GitHubID, &u.Username, &u.AvatarURL, &u.Email, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepo) GetByGitHubID(ctx context.Context, githubID string) (*models.User, error) {
	query := `SELECT id, github_id, username, avatar_url, email, created_at FROM users WHERE github_id = $1`
	var u models.User
	err := r.pool.QueryRow(ctx, query, githubID).Scan(&u.ID, &u.GitHubID, &u.Username, &u.AvatarURL, &u.Email, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by github id: %w", err)
	}
	return &u, nil
}

func (r *PostgresUserRepo) Create(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (github_id, username, avatar_url, email, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, user.GitHubID, user.Username, user.AvatarURL, user.Email).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (r *PostgresUserRepo) UpsertByGitHub(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (github_id, username, avatar_url, email, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (github_id) DO UPDATE SET
			username = EXCLUDED.username,
			avatar_url = EXCLUDED.avatar_url,
			email = EXCLUDED.email
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query, user.GitHubID, user.Username, user.AvatarURL, user.Email).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert user: %w", err)
	}
	return user, nil
}

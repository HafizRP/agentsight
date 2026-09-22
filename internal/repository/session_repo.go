package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"agentsight/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresSessionRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresSessionRepo(pool *pgxpool.Pool) *PostgresSessionRepo {
	return &PostgresSessionRepo{pool: pool}
}

func (r *PostgresSessionRepo) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (token, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err := r.pool.Exec(ctx, query, session.Token, session.UserID, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

func (r *PostgresSessionRepo) GetSession(ctx context.Context, token string) (*models.Session, error) {
	query := `SELECT token, user_id, expires_at, created_at FROM sessions WHERE token = $1 AND expires_at > NOW()`
	var s models.Session
	err := r.pool.QueryRow(ctx, query, token).Scan(&s.Token, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("session not found or expired")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &s, nil
}

func (r *PostgresSessionRepo) DeleteSession(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func (r *PostgresSessionRepo) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at <= $1`
	_, err := r.pool.Exec(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}

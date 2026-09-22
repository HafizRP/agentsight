package repository

import (
	"context"
	"fmt"

	"agentsight/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresScrapeLogRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresScrapeLogRepo(pool *pgxpool.Pool) *PostgresScrapeLogRepo {
	return &PostgresScrapeLogRepo{pool: pool}
}

func (r *PostgresScrapeLogRepo) Create(ctx context.Context, log *models.ScrapeLog) error {
	query := `
		INSERT INTO scrape_logs (source_id, status, skills_found, skills_new, skills_updated, error_message, duration_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query,
		log.SourceID, log.Status, log.SkillsFound, log.SkillsNew, log.SkillsUpdated, log.ErrorMessage, log.DurationMS,
	).Scan(&log.ID, &log.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create scrape log: %w", err)
	}
	return nil
}

func (r *PostgresScrapeLogRepo) ListRecent(ctx context.Context, limit int) ([]models.ScrapeLog, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, source_id, status, skills_found, skills_new, skills_updated, error_message, duration_ms, created_at
		FROM scrape_logs
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list scrape logs: %w", err)
	}
	defer rows.Close()

	var logs []models.ScrapeLog
	for rows.Next() {
		var l models.ScrapeLog
		err := rows.Scan(
			&l.ID, &l.SourceID, &l.Status, &l.SkillsFound, &l.SkillsNew, &l.SkillsUpdated, &l.ErrorMessage, &l.DurationMS, &l.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan scrape log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

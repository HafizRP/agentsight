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

type PostgresSourceRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresSourceRepo(pool *pgxpool.Pool) *PostgresSourceRepo {
	return &PostgresSourceRepo{pool: pool}
}

func (r *PostgresSourceRepo) Create(ctx context.Context, source *models.Source) (*models.Source, error) {
	query := `
		INSERT INTO sources (name, url, source_type, scrape_strategy, is_active, last_scraped_at, scrape_interval_hours, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING id, created_at
	`
	err := r.pool.QueryRow(ctx, query,
		source.Name, source.URL, source.SourceType, source.ScrapeStrategy,
		source.IsActive, source.LastScrapedAt, source.ScrapeIntervalHours,
	).Scan(&source.ID, &source.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}
	return source, nil
}

func (r *PostgresSourceRepo) GetByID(ctx context.Context, id int) (*models.Source, error) {
	query := `
		SELECT id, name, url, source_type, scrape_strategy, is_active, last_scraped_at, scrape_interval_hours, created_at
		FROM sources WHERE id = $1
	`
	var s models.Source
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.URL, &s.SourceType, &s.ScrapeStrategy,
		&s.IsActive, &s.LastScrapedAt, &s.ScrapeIntervalHours, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("source not found")
		}
		return nil, fmt.Errorf("failed to get source: %w", err)
	}
	return &s, nil
}

func (r *PostgresSourceRepo) List(ctx context.Context) ([]models.Source, error) {
	query := `
		SELECT id, name, url, source_type, scrape_strategy, is_active, last_scraped_at, scrape_interval_hours, created_at
		FROM sources
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		var s models.Source
		err := rows.Scan(
			&s.ID, &s.Name, &s.URL, &s.SourceType, &s.ScrapeStrategy,
			&s.IsActive, &s.LastScrapedAt, &s.ScrapeIntervalHours, &s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		sources = append(sources, s)
	}
	return sources, nil
}

func (r *PostgresSourceRepo) ListActive(ctx context.Context) ([]models.Source, error) {
	query := `
		SELECT id, name, url, source_type, scrape_strategy, is_active, last_scraped_at, scrape_interval_hours, created_at
		FROM sources
		WHERE is_active = true
		ORDER BY id ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list active sources: %w", err)
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		var s models.Source
		err := rows.Scan(
			&s.ID, &s.Name, &s.URL, &s.SourceType, &s.ScrapeStrategy,
			&s.IsActive, &s.LastScrapedAt, &s.ScrapeIntervalHours, &s.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		sources = append(sources, s)
	}
	return sources, nil
}

func (r *PostgresSourceRepo) Update(ctx context.Context, source *models.Source) error {
	query := `
		UPDATE sources SET
			name = $1, url = $2, source_type = $3, scrape_strategy = $4,
			is_active = $5, last_scraped_at = $6, scrape_interval_hours = $7
		WHERE id = $8
	`
	tag, err := r.pool.Exec(ctx, query,
		source.Name, source.URL, source.SourceType, source.ScrapeStrategy,
		source.IsActive, source.LastScrapedAt, source.ScrapeIntervalHours,
		source.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update source: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("source not found")
	}
	return nil
}

func (r *PostgresSourceRepo) UpdateLastScraped(ctx context.Context, id int, t time.Time) error {
	query := `UPDATE sources SET last_scraped_at = $1 WHERE id = $2`
	tag, err := r.pool.Exec(ctx, query, t, id)
	if err != nil {
		return fmt.Errorf("failed to update last scraped: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("source not found")
	}
	return nil
}

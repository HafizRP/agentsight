package repository

import (
	"context"
	"fmt"

	"agentsight/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresBookmarkRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresBookmarkRepo(pool *pgxpool.Pool) *PostgresBookmarkRepo {
	return &PostgresBookmarkRepo{pool: pool}
}

func (r *PostgresBookmarkRepo) Toggle(ctx context.Context, userID, skillID int) (bool, error) {
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM bookmarks WHERE user_id = $1 AND skill_id = $2)`
	err := r.pool.QueryRow(ctx, checkQuery, userID, skillID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check bookmark: %w", err)
	}

	if exists {
		deleteQuery := `DELETE FROM bookmarks WHERE user_id = $1 AND skill_id = $2`
		_, err := r.pool.Exec(ctx, deleteQuery, userID, skillID)
		if err != nil {
			return false, fmt.Errorf("failed to remove bookmark: %w", err)
		}
		return false, nil
	}

	insertQuery := `INSERT INTO bookmarks (user_id, skill_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING`
	_, err = r.pool.Exec(ctx, insertQuery, userID, skillID)
	if err != nil {
		return false, fmt.Errorf("failed to add bookmark: %w", err)
	}
	return true, nil
}

func (r *PostgresBookmarkRepo) IsBookmarked(ctx context.Context, userID, skillID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM bookmarks WHERE user_id = $1 AND skill_id = $2)`
	err := r.pool.QueryRow(ctx, query, userID, skillID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if bookmarked: %w", err)
	}
	return exists, nil
}

func (r *PostgresBookmarkRepo) GetBookmarkedSkillIDs(ctx context.Context, userID int) (map[int]bool, error) {
	query := `SELECT skill_id FROM bookmarks WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmarked skill ids: %w", err)
	}
	defer rows.Close()

	res := make(map[int]bool)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			res[id] = true
		}
	}
	return res, nil
}

func (r *PostgresBookmarkRepo) ListByUser(ctx context.Context, userID int, limit, offset int) ([]models.Skill, int, error) {
	countQuery := `SELECT COUNT(*) FROM bookmarks WHERE user_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user bookmarks: %w", err)
	}

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT s.id, s.name, s.slug, s.description, s.platform, s.category, s.subcategory,
		       s.source_url, s.repo_url, s.repo_owner, s.repo_name, s.stars_count, s.forks_count, s.stars_velocity,
		       s.content_raw, s.content_preview, s.install_snippet, s.file_path, s.language, s.tags,
		       s.last_synced_at, s.created_at, s.updated_at
		FROM skills s
		INNER JOIN bookmarks b ON b.skill_id = s.id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list user bookmarks: %w", err)
	}
	defer rows.Close()

	skills, err := scanSkills(rows)
	if err != nil {
		return nil, 0, err
	}
	return skills, total, nil
}

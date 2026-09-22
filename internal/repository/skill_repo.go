package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"agentsight/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresSkillRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresSkillRepo(pool *pgxpool.Pool) *PostgresSkillRepo {
	return &PostgresSkillRepo{pool: pool}
}

func (r *PostgresSkillRepo) Create(ctx context.Context, skill *models.Skill) (*models.Skill, error) {
	query := `
		INSERT INTO skills (
			name, slug, description, platform, category, subcategory,
			source_url, repo_url, repo_owner, repo_name,
			stars_count, forks_count, stars_velocity,
			content_raw, content_preview, install_snippet, file_path, language, tags,
			last_synced_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19,
			$20, NOW(), NOW()
		)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query,
		skill.Name, skill.Slug, skill.Description, string(skill.Platform), string(skill.Category), skill.Subcategory,
		skill.SourceURL, skill.RepoURL, skill.RepoOwner, skill.RepoName,
		skill.StarsCount, skill.ForksCount, skill.StarsVelocity,
		skill.ContentRaw, skill.ContentPreview, skill.InstallSnippet, skill.FilePath, skill.Language, skill.Tags,
		skill.LastSyncedAt,
	).Scan(&skill.ID, &skill.CreatedAt, &skill.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create skill: %w", err)
	}
	return skill, nil
}

func (r *PostgresSkillRepo) GetByID(ctx context.Context, id int) (*models.Skill, error) {
	query := `
		SELECT id, name, slug, description, platform, category, subcategory,
		       source_url, repo_url, repo_owner, repo_name, stars_count, forks_count, stars_velocity,
		       content_raw, content_preview, install_snippet, file_path, language, tags,
		       last_synced_at, created_at, updated_at
		FROM skills
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	return scanSkill(row)
}

func (r *PostgresSkillRepo) GetBySlug(ctx context.Context, slug string) (*models.Skill, error) {
	query := `
		SELECT id, name, slug, description, platform, category, subcategory,
		       source_url, repo_url, repo_owner, repo_name, stars_count, forks_count, stars_velocity,
		       content_raw, content_preview, install_snippet, file_path, language, tags,
		       last_synced_at, created_at, updated_at
		FROM skills
		WHERE slug = $1
	`
	row := r.pool.QueryRow(ctx, query, slug)
	return scanSkill(row)
}

func (r *PostgresSkillRepo) Update(ctx context.Context, skill *models.Skill) error {
	query := `
		UPDATE skills SET
			name = $1, description = $2, platform = $3, category = $4, subcategory = $5,
			source_url = $6, repo_url = $7, repo_owner = $8, repo_name = $9,
			stars_count = $10, forks_count = $11, stars_velocity = $12,
			content_raw = $13, content_preview = $14, install_snippet = $15,
			file_path = $16, language = $17, tags = $18, last_synced_at = $19,
			updated_at = NOW()
		WHERE id = $20
	`
	tag, err := r.pool.Exec(ctx, query,
		skill.Name, skill.Description, string(skill.Platform), string(skill.Category), skill.Subcategory,
		skill.SourceURL, skill.RepoURL, skill.RepoOwner, skill.RepoName,
		skill.StarsCount, skill.ForksCount, skill.StarsVelocity,
		skill.ContentRaw, skill.ContentPreview, skill.InstallSnippet,
		skill.FilePath, skill.Language, skill.Tags, skill.LastSyncedAt,
		skill.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update skill: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("skill not found")
	}
	return nil
}

func (r *PostgresSkillRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM skills WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete skill: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("skill not found")
	}
	return nil
}

func (r *PostgresSkillRepo) List(ctx context.Context, filter models.SkillFilter) ([]models.Skill, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.Platform != "" {
		conditions = append(conditions, fmt.Sprintf("platform = $%d", argIdx))
		args = append(args, filter.Platform)
		argIdx++
	}
	if filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, filter.Category)
		argIdx++
	}
	if filter.Subcategory != "" {
		conditions = append(conditions, fmt.Sprintf("subcategory = $%d", argIdx))
		args = append(args, filter.Subcategory)
		argIdx++
	}
	if filter.Language != "" {
		conditions = append(conditions, fmt.Sprintf("language ILIKE $%d", argIdx))
		args = append(args, filter.Language)
		argIdx++
	}
	if filter.Tag != "" {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(tags)", argIdx))
		args = append(args, filter.Tag)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM skills %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count skills: %w", err)
	}

	orderBy := "ORDER BY stars_count DESC"
	switch filter.Sort {
	case "trending":
		orderBy = "ORDER BY stars_velocity DESC, stars_count DESC"
	case "newest":
		orderBy = "ORDER BY created_at DESC"
	case "stars":
		orderBy = "ORDER BY stars_count DESC"
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (filter.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, name, slug, description, platform, category, subcategory,
		       source_url, repo_url, repo_owner, repo_name, stars_count, forks_count, stars_velocity,
		       content_raw, content_preview, install_snippet, file_path, language, tags,
		       last_synced_at, created_at, updated_at
		FROM skills
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIdx, argIdx+1)

	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query skills: %w", err)
	}
	defer rows.Close()

	skills, err := scanSkills(rows)
	if err != nil {
		return nil, 0, err
	}
	return skills, total, nil
}

func (r *PostgresSkillRepo) Search(ctx context.Context, params models.SearchParams) ([]models.Skill, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	var rankSelect string
	var orderBy string

	if strings.TrimSpace(params.Query) != "" {
		conditions = append(conditions, fmt.Sprintf(`(
			to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, '') || ' ' || coalesce(content_raw, '')) @@ plainto_tsquery('english', $%d)
			OR name ILIKE '%%' || $%d || '%%'
			OR description ILIKE '%%' || $%d || '%%'
			OR $%d = ANY(tags)
		)`, argIdx, argIdx, argIdx, argIdx))
		args = append(args, params.Query)
		rankSelect = fmt.Sprintf(", ts_rank(to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, '') || ' ' || coalesce(content_raw, '')), plainto_tsquery('english', $%d)) AS rank", argIdx)
		argIdx++
	}

	if params.Platform != "" {
		conditions = append(conditions, fmt.Sprintf("platform = $%d", argIdx))
		args = append(args, params.Platform)
		argIdx++
	}
	if params.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, params.Category)
		argIdx++
	}
	if params.Subcategory != "" {
		conditions = append(conditions, fmt.Sprintf("subcategory = $%d", argIdx))
		args = append(args, params.Subcategory)
		argIdx++
	}
	if params.Language != "" {
		conditions = append(conditions, fmt.Sprintf("language ILIKE $%d", argIdx))
		args = append(args, params.Language)
		argIdx++
	}
	if params.Tag != "" {
		conditions = append(conditions, fmt.Sprintf("$%d = ANY(tags)", argIdx))
		args = append(args, params.Tag)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM skills %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	switch params.Sort {
	case "stars":
		orderBy = "ORDER BY stars_count DESC"
	case "trending":
		orderBy = "ORDER BY stars_velocity DESC, stars_count DESC"
	case "newest":
		orderBy = "ORDER BY created_at DESC"
	case "relevance":
		if rankSelect != "" {
			orderBy = "ORDER BY rank DESC, stars_count DESC"
		} else {
			orderBy = "ORDER BY stars_count DESC"
		}
	default:
		if rankSelect != "" {
			orderBy = "ORDER BY rank DESC, stars_count DESC"
		} else {
			orderBy = "ORDER BY stars_count DESC"
		}
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (params.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, name, slug, description, platform, category, subcategory,
		       source_url, repo_url, repo_owner, repo_name, stars_count, forks_count, stars_velocity,
		       content_raw, content_preview, install_snippet, file_path, language, tags,
		       last_synced_at, created_at, updated_at
		       %s
		FROM skills
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, rankSelect, whereClause, orderBy, argIdx, argIdx+1)

	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search: %w", err)
	}
	defer rows.Close()

	var skills []models.Skill
	for rows.Next() {
		var s models.Skill
		var platformStr, categoryStr string
		var dest []interface{}
		dest = []interface{}{
			&s.ID, &s.Name, &s.Slug, &s.Description, &platformStr, &categoryStr, &s.Subcategory,
			&s.SourceURL, &s.RepoURL, &s.RepoOwner, &s.RepoName, &s.StarsCount, &s.ForksCount, &s.StarsVelocity,
			&s.ContentRaw, &s.ContentPreview, &s.InstallSnippet, &s.FilePath, &s.Language, &s.Tags,
			&s.LastSyncedAt, &s.CreatedAt, &s.UpdatedAt,
		}
		if rankSelect != "" {
			var rank float64
			dest = append(dest, &rank)
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, 0, fmt.Errorf("failed to scan search row: %w", err)
		}
		s.Platform = models.PlatformType(platformStr)
		s.Category = models.CategoryType(categoryStr)
		skills = append(skills, s)
	}

	return skills, total, nil
}

func (r *PostgresSkillRepo) GetTrending(ctx context.Context, limit int) ([]models.Skill, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `
		SELECT id, name, slug, description, platform, category, subcategory,
		       source_url, repo_url, repo_owner, repo_name, stars_count, forks_count, stars_velocity,
		       content_raw, content_preview, install_snippet, file_path, language, tags,
		       last_synced_at, created_at, updated_at
		FROM skills
		ORDER BY stars_velocity DESC, stars_count DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query trending: %w", err)
	}
	defer rows.Close()
	return scanSkills(rows)
}

func (r *PostgresSkillRepo) ListByCategory(ctx context.Context, category string, limit, offset int) ([]models.Skill, int, error) {
	filter := models.SkillFilter{
		Category: category,
		Limit:    limit,
		Page:     (offset / limit) + 1,
		Sort:     "stars",
	}
	return r.List(ctx, filter)
}

func (r *PostgresSkillRepo) ListByPlatform(ctx context.Context, platform string, limit, offset int) ([]models.Skill, int, error) {
	filter := models.SkillFilter{
		Platform: platform,
		Limit:    limit,
		Page:     (offset / limit) + 1,
		Sort:     "stars",
	}
	return r.List(ctx, filter)
}

func (r *PostgresSkillRepo) GetRelated(ctx context.Context, skill *models.Skill, limit int) ([]models.Skill, error) {
	if limit <= 0 {
		limit = 4
	}
	query := `
		SELECT id, name, slug, description, platform, category, subcategory,
		       source_url, repo_url, repo_owner, repo_name, stars_count, forks_count, stars_velocity,
		       content_raw, content_preview, install_snippet, file_path, language, tags,
		       last_synced_at, created_at, updated_at
		FROM skills
		WHERE id != $1
		  AND (category = $2 OR platform = $3 OR tags && $4)
		ORDER BY stars_count DESC
		LIMIT $5
	`
	rows, err := r.pool.Query(ctx, query, skill.ID, string(skill.Category), string(skill.Platform), skill.Tags, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get related skills: %w", err)
	}
	defer rows.Close()
	return scanSkills(rows)
}

func (r *PostgresSkillRepo) GetStats(ctx context.Context) (*models.Stats, error) {
	stats := &models.Stats{
		PlatformCounts: make(map[string]int),
		CategoryCounts: make(map[string]int),
	}

	totalQuery := `
		SELECT COUNT(*), COALESCE(SUM(stars_count), 0)
		FROM skills
	`
	err := r.pool.QueryRow(ctx, totalQuery).Scan(&stats.TotalSkills, &stats.TotalStars)
	if err != nil {
		return nil, fmt.Errorf("failed to get total stats: %w", err)
	}

	sourcesQuery := `SELECT COUNT(*) FROM sources WHERE is_active = true`
	_ = r.pool.QueryRow(ctx, sourcesQuery).Scan(&stats.TotalSources)

	pRows, err := r.pool.Query(ctx, `SELECT platform, COUNT(*) FROM skills GROUP BY platform`)
	if err == nil {
		defer pRows.Close()
		for pRows.Next() {
			var p string
			var count int
			if err := pRows.Scan(&p, &count); err == nil {
				stats.PlatformCounts[p] = count
			}
		}
	}

	cRows, err := r.pool.Query(ctx, `SELECT category, COUNT(*) FROM skills GROUP BY category`)
	if err == nil {
		defer cRows.Close()
		for cRows.Next() {
			var c string
			var count int
			if err := cRows.Scan(&c, &count); err == nil {
				stats.CategoryCounts[c] = count
			}
		}
	}

	return stats, nil
}

func scanSkill(row pgx.Row) (*models.Skill, error) {
	var s models.Skill
	var platformStr, categoryStr string
	err := row.Scan(
		&s.ID, &s.Name, &s.Slug, &s.Description, &platformStr, &categoryStr, &s.Subcategory,
		&s.SourceURL, &s.RepoURL, &s.RepoOwner, &s.RepoName, &s.StarsCount, &s.ForksCount, &s.StarsVelocity,
		&s.ContentRaw, &s.ContentPreview, &s.InstallSnippet, &s.FilePath, &s.Language, &s.Tags,
		&s.LastSyncedAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("skill not found")
		}
		return nil, fmt.Errorf("failed to scan skill: %w", err)
	}
	s.Platform = models.PlatformType(platformStr)
	s.Category = models.CategoryType(categoryStr)
	return &s, nil
}

func scanSkills(rows pgx.Rows) ([]models.Skill, error) {
	var skills []models.Skill
	for rows.Next() {
		var s models.Skill
		var platformStr, categoryStr string
		err := rows.Scan(
			&s.ID, &s.Name, &s.Slug, &s.Description, &platformStr, &categoryStr, &s.Subcategory,
			&s.SourceURL, &s.RepoURL, &s.RepoOwner, &s.RepoName, &s.StarsCount, &s.ForksCount, &s.StarsVelocity,
			&s.ContentRaw, &s.ContentPreview, &s.InstallSnippet, &s.FilePath, &s.Language, &s.Tags,
			&s.LastSyncedAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan skill row: %w", err)
		}
		s.Platform = models.PlatformType(platformStr)
		s.Category = models.CategoryType(categoryStr)
		skills = append(skills, s)
	}
	return skills, nil
}

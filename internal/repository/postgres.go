package repository

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"agentsight/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("invalid db connection string: %w", err)
	}
	cfg.MaxConns = 25
	cfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationFiles := []string{
		"001_init_schema.sql",
		"002_seed_skills.sql",
		"003_seed_sources.sql",
	}

	for _, filename := range migrationFiles {
		content, err := migrations.FS.ReadFile(filename)
		if err != nil {
			slog.Warn("could not read migration file from embed", "file", filename, "err", err)
			continue
		}

		sqlStr := string(content)
		if strings.TrimSpace(sqlStr) == "" {
			continue
		}

		slog.Info("executing migration", "file", filename)
		_, err = pool.Exec(ctx, sqlStr)
		if err != nil {
			return fmt.Errorf("migration %s failed: %w", filename, err)
		}
	}

	slog.Info("database migrations completed successfully")
	return nil
}

func NewPostgresRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		Skills:     NewPostgresSkillRepo(pool),
		Sources:    NewPostgresSourceRepo(pool),
		Users:      NewPostgresUserRepo(pool),
		Bookmarks:  NewPostgresBookmarkRepo(pool),
		ScrapeLogs: NewPostgresScrapeLogRepo(pool),
		Sessions:   NewPostgresSessionRepo(pool),
	}
}

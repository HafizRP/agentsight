-- Migrations: 001_init_schema.sql
-- Database schema for AgentSight

-- Enums with idempotency guards
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'platform_type') THEN
        CREATE TYPE platform_type AS ENUM ('cursor', 'claude', 'gemini', 'mcp', 'copilot', 'generic');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'category_type') THEN
        CREATE TYPE category_type AS ENUM ('rules', 'skills', 'mcp_servers', 'prompts', 'frameworks', 'tools');
    END IF;
END $$;

-- skills: the core entity — one record per discovered agent skill
CREATE TABLE IF NOT EXISTS skills (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) UNIQUE NOT NULL,
  description TEXT,
  platform platform_type NOT NULL,
  category category_type NOT NULL,
  subcategory VARCHAR(100),
  source_url VARCHAR(500) NOT NULL,
  repo_url VARCHAR(500),
  repo_owner VARCHAR(100),
  repo_name VARCHAR(200),
  stars_count INT DEFAULT 0,
  forks_count INT DEFAULT 0,
  stars_velocity INT DEFAULT 0,
  content_raw TEXT,
  content_preview VARCHAR(500),
  install_snippet TEXT,
  file_path VARCHAR(500),
  language VARCHAR(50),
  tags TEXT[] DEFAULT '{}',
  last_synced_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- sources: scraper targets registry
CREATE TABLE IF NOT EXISTS sources (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  url VARCHAR(500) NOT NULL,
  source_type VARCHAR(50) NOT NULL,
  scrape_strategy VARCHAR(50),
  is_active BOOLEAN DEFAULT true,
  last_scraped_at TIMESTAMPTZ,
  scrape_interval_hours INT DEFAULT 24,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- users: for bookmarks & personalization
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  github_id VARCHAR(50) UNIQUE,
  username VARCHAR(100) NOT NULL,
  avatar_url VARCHAR(500),
  email VARCHAR(255),
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- sessions: server-side session management
CREATE TABLE IF NOT EXISTS sessions (
  token VARCHAR(128) PRIMARY KEY,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- bookmarks: user-skill favorites
CREATE TABLE IF NOT EXISTS bookmarks (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id) ON DELETE CASCADE,
  skill_id INT REFERENCES skills(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, skill_id)
);

-- scrape_logs: audit trail for scraper runs
CREATE TABLE IF NOT EXISTS scrape_logs (
  id SERIAL PRIMARY KEY,
  source_id INT REFERENCES sources(id) ON DELETE SET NULL,
  status VARCHAR(20),
  skills_found INT DEFAULT 0,
  skills_new INT DEFAULT 0,
  skills_updated INT DEFAULT 0,
  error_message TEXT,
  duration_ms INT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_skills_slug ON skills(slug);
CREATE INDEX IF NOT EXISTS idx_skills_platform ON skills(platform);
CREATE INDEX IF NOT EXISTS idx_skills_category ON skills(category);
CREATE INDEX IF NOT EXISTS idx_skills_subcategory ON skills(subcategory);
CREATE INDEX IF NOT EXISTS idx_skills_stars_count ON skills(stars_count DESC);
CREATE INDEX IF NOT EXISTS idx_skills_stars_velocity ON skills(stars_velocity DESC);
CREATE INDEX IF NOT EXISTS idx_skills_tags ON skills USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_skills_fts ON skills USING GIN (
  to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, '') || ' ' || coalesce(content_raw, ''))
);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX IF NOT EXISTS idx_bookmarks_skill_id ON bookmarks(skill_id);
CREATE INDEX IF NOT EXISTS idx_scrape_logs_source_id ON scrape_logs(source_id);

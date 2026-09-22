-- Migrations: 003_seed_sources.sql
-- Seed scraper targets

INSERT INTO sources (name, url, source_type, scrape_strategy, scrape_interval_hours) VALUES
('Awesome Cursor Rules', 'https://github.com/PatrickJS/awesome-cursorrules', 'github_repo', 'api', 24),
('Awesome MCP Servers', 'https://github.com/punkpeye/awesome-mcp-servers', 'github_repo', 'api', 24),
('Official MCP Servers', 'https://github.com/modelcontextprotocol/servers', 'github_repo', 'api', 48),
('AGENTS.md Spec', 'https://github.com/agentsmd/agents.md', 'github_repo', 'api', 48),
('Cursor Directory', 'https://cursor.directory', 'website', 'chromedp', 24),
('MCP Servers Org', 'https://mcpservers.org', 'website', 'chromedp', 24),
('GitHub Code Search - cursorrules', 'https://api.github.com/search/code?q=filename:.cursorrules', 'github_search', 'api', 72),
('GitHub Code Search - AGENTS.md', 'https://api.github.com/search/code?q=filename:AGENTS.md', 'github_search', 'api', 72)
ON CONFLICT DO NOTHING;

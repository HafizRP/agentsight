package scraper_test

import (
	"testing"

	"agentsight/internal/scraper"
)

const sampleAwesomeMarkdown = `# Awesome AI Agent Skills & Tools

Curated list of AI agent skills, prompts, and frameworks.

## Table of Contents
- [Frameworks](#frameworks)
- [MCP Servers](#mcp-servers)

## Frameworks

### Python Multi-Agent Systems
- [LangGraph](https://github.com/langchain-ai/langgraph) - Build resilient language agents as graphs.
* [AutoGen](https://github.com/microsoft/autogen): Multi-agent conversation framework.
- **[CrewAI](https://github.com/crewAIInc/crewAI)** — Orchestrating role-playing autonomous AI agents.
1. [DSPy](https://github.com/stanfordnlp/dspy) – Framework for programming with foundation models.

### Rust High Performance
- [Rig](https://github.com/0xPlaygrounds/rig) - Open source library for building LLM applications in Rust.

## MCP Servers
| Server | Description | Repository |
|---|---|---|
| [PostgreSQL](https://github.com/modelcontextprotocol/servers) | PostgreSQL database context provider | https://github.com/modelcontextprotocol/servers |
| [Playwright](https://github.com/executeautomation/mcp-playwright) | Browser automation and end-to-end testing | https://github.com/executeautomation/mcp-playwright |
| SQLite Explorer | Embedded SQLite inspection server | https://github.com/example/sqlite-mcp |

## Ignored Section
<!-- Comment block -->
Check out contributing guidelines.
`

func TestAwesomeListParser_Parse_BulletLists(t *testing.T) {
	parser := scraper.NewAwesomeListParser()
	entries := parser.Parse(sampleAwesomeMarkdown)

	if len(entries) == 0 {
		t.Fatalf("expected parsed entries, got 0")
	}

	// Verify LangGraph entry
	var langGraph *scraper.AwesomeEntry
	for i := range entries {
		if entries[i].Name == "LangGraph" {
			langGraph = &entries[i]
			break
		}
	}

	if langGraph == nil {
		t.Fatalf("expected to find LangGraph entry")
	}
	if langGraph.Category != "Frameworks" {
		t.Errorf("expected Category 'Frameworks', got %q", langGraph.Category)
	}
	if langGraph.Subcategory != "Python Multi-Agent Systems" {
		t.Errorf("expected Subcategory 'Python Multi-Agent Systems', got %q", langGraph.Subcategory)
	}
	if langGraph.URL != "https://github.com/langchain-ai/langgraph" {
		t.Errorf("expected LangGraph GitHub URL, got %q", langGraph.URL)
	}
	if langGraph.Description != "Build resilient language agents as graphs." {
		t.Errorf("expected clean description, got %q", langGraph.Description)
	}

	// Verify AutoGen with colon delimiter
	var autogen *scraper.AwesomeEntry
	for i := range entries {
		if entries[i].Name == "AutoGen" {
			autogen = &entries[i]
			break
		}
	}
	if autogen == nil {
		t.Fatalf("expected to find AutoGen entry")
	}
	if autogen.Description != "Multi-agent conversation framework." {
		t.Errorf("expected AutoGen description, got %q", autogen.Description)
	}

	// Verify CrewAI with bold markdown and em-dash delimiter
	var crewAI *scraper.AwesomeEntry
	for i := range entries {
		if entries[i].Name == "CrewAI" {
			crewAI = &entries[i]
			break
		}
	}
	if crewAI == nil {
		t.Fatalf("expected to find CrewAI entry")
	}
	if crewAI.Description != "Orchestrating role-playing autonomous AI agents." {
		t.Errorf("expected CrewAI description, got %q", crewAI.Description)
	}

	// Verify DSPy with numeric bullet and en-dash delimiter
	var dspy *scraper.AwesomeEntry
	for i := range entries {
		if entries[i].Name == "DSPy" {
			dspy = &entries[i]
			break
		}
	}
	if dspy == nil {
		t.Fatalf("expected to find DSPy entry")
	}
	if dspy.Description != "Framework for programming with foundation models." {
		t.Errorf("expected DSPy description, got %q", dspy.Description)
	}
}

func TestAwesomeListParser_Parse_Tables(t *testing.T) {
	parser := scraper.NewAwesomeListParser()
	entries := parser.Parse(sampleAwesomeMarkdown)

	// Verify PostgreSQL from table
	var postgres *scraper.AwesomeEntry
	for i := range entries {
		if entries[i].Name == "PostgreSQL" {
			postgres = &entries[i]
			break
		}
	}

	if postgres == nil {
		t.Fatalf("expected to find PostgreSQL table entry")
	}
	if postgres.Category != "MCP Servers" {
		t.Errorf("expected Category 'MCP Servers', got %q", postgres.Category)
	}
	if postgres.URL != "https://github.com/modelcontextprotocol/servers" {
		t.Errorf("expected PostgreSQL repo URL, got %q", postgres.URL)
	}
	if postgres.Description != "PostgreSQL database context provider" {
		t.Errorf("expected PostgreSQL description, got %q", postgres.Description)
	}

	// Verify SQLite Explorer where link is in the repository column
	var sqlite *scraper.AwesomeEntry
	for i := range entries {
		if entries[i].Name == "SQLite Explorer" {
			sqlite = &entries[i]
			break
		}
	}
	if sqlite == nil {
		t.Fatalf("expected to find SQLite Explorer entry")
	}
	if sqlite.URL != "https://github.com/example/sqlite-mcp" {
		t.Errorf("expected SQLite Explorer URL, got %q", sqlite.URL)
	}
}

func TestAwesomeListParser_ParseToRawSkills(t *testing.T) {
	parser := scraper.NewAwesomeListParser()
	skills := parser.ParseToRawSkills(sampleAwesomeMarkdown, "defaultOwner", "defaultRepo", 5000, 450)

	if len(skills) < 5 {
		t.Fatalf("expected at least 5 raw skills, got %d", len(skills))
	}

	var langGraphSkill *scraper.RawSkill
	for i := range skills {
		if skills[i].Name == "LangGraph" {
			langGraphSkill = &skills[i]
			break
		}
	}

	if langGraphSkill == nil {
		t.Fatalf("expected to find LangGraph RawSkill")
	}
	if langGraphSkill.RepoOwner != "langchain-ai" {
		t.Errorf("expected RepoOwner 'langchain-ai', got %q", langGraphSkill.RepoOwner)
	}
	if langGraphSkill.RepoName != "langgraph" {
		t.Errorf("expected RepoName 'langgraph', got %q", langGraphSkill.RepoName)
	}
	if langGraphSkill.StarsCount != 5000 {
		t.Errorf("expected StarsCount 5000, got %d", langGraphSkill.StarsCount)
	}
	if langGraphSkill.ForksCount != 450 {
		t.Errorf("expected ForksCount 450, got %d", langGraphSkill.ForksCount)
	}
}

func TestAwesomeListParser_EmptyAndCorruptInput(t *testing.T) {
	parser := scraper.NewAwesomeListParser()

	// Empty string
	entries := parser.Parse("")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for empty input, got %d", len(entries))
	}

	// Random text with no links
	entries = parser.Parse("Just some random text\nwithout any markdown links or tables.")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for text without links, got %d", len(entries))
	}

	// Anchors only
	entries = parser.Parse("- [Back to Top](#top)\n- [Overview](#overview)")
	if len(entries) != 0 {
		t.Errorf("expected anchors to be ignored, got %d", len(entries))
	}
}

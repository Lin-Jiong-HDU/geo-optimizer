# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **B2B GEO (Generative Engine Optimization)** engine implemented as a Go library package. It optimizes content to improve visibility and citation rates in AI search engines (ChatGPT, Perplexity, Google AI) through LLM-driven content transformation.

**Target Users**: GEO service providers who integrate this package into their backend systems to offer GEO optimization services to enterprise clients.

## Architecture

### Package Structure

```
pkg/
├── optimizer/         # Core optimization engine
│   ├── optimizer.go   # Main optimizer orchestrator
│   ├── strategies/    # Individual strategy implementations
│   └── enhancer/      # Content enhancement utilities
├── llm/               # LLM abstraction layer
│   ├── client.go      # Unified LLM client interface
│   ├── provider.go    # Provider-specific implementations
│   ├── prompts/       # Prompt templates and builders
│   └── parser.go      # LLM response parsing
├── models/            # Data models (Request/Response/Enterprise/Content)
├── config/            # Configuration management
│   └── ai_profiles.go # Predefined AI platform preferences
└── analyzer/          # Content analysis and scoring
    └── scorer.go      # GEO scoring system
```

### Key Design Patterns

1. **LLM Abstraction**: `llm.LLMClient` interface supports multiple providers (OpenAI, Anthropic, Azure, Google, custom endpoints). Implementations handle provider-specific differences while exposing a unified API.

2. **Strategy Pattern**: `optimizer/strategies/` implements discrete optimization strategies (structure, schema, answer-first, authority, FAQ). Each strategy can be applied independently or combined.

3. **Request-Response Model**: All models in `pkg/models/` use JSON serialization for easy integration with other backend services.

4. **Multi-Tenant Architecture**: Designed for enterprise isolation through `EnterpriseInfo` in requests, allowing service providers to handle multiple clients.

## Core Data Flow

1. **Input**: `OptimizationRequest` containing content, enterprise info, competitors, and AI preferences
2. **Analysis**: `analyzer.Scorer` evaluates original content (GeoScore)
3. **Optimization**: `optimizer.Optimizer` applies selected strategies using LLM
4. **Parsing**: `llm.parser` extracts structured optimization from LLM response
5. **Output**: `OptimizationResponse` with optimized content, schema markup, and metrics

## Important Concepts

### GEO vs SEO

- **SEO**: Optimizes for traditional search engine rankings (Google, Bing)
- **GEO**: Optimizes for AI engine citations and recommendations (ChatGPT, Perplexity, Google AI)
- Key difference: Structured, authoritative, answer-first content that AI models can easily reference

### Optimization Strategies

- `structure`: Adds clear heading hierarchy, bullet points, organized sections
- `schema`: Generates JSON-LD structured data markup
- `answer_first`: Moves key conclusions to the beginning
- `authority`: Enhances with citations, sources, credentials
- `faq`: Generates FAQ sections for common queries

### AI Platform Preferences

The system supports platform-specific customization via `AIPreferences`:

- **ChatGPT**: Professional tone, informative citation style
- **Perplexity**: Concise, question-oriented headings
- **Google AI**: Structured, schema-heavy content

## Development Commands

```bash
# Initialize module
go mod init github.com/Lin-Jiong-HDU/geo-optimizer

# Build the package
go build ./...

# Run tests
go test ./... -v

# Run specific package tests
go test ./pkg/optimizer -v

# Run single test
go test ./pkg/models -run TestOptimizationRequest

# Lint (requires golangci-lint)
golangci-lint run

# Format code
go fmt ./...

# Tidy dependencies
go mod tidy
```

## Module Information

- **Module Path**: `github.com/Lin-Jiong-HDU/geo-optimizer`
- **Go Version**: 1.25.5
- **Package Import**: `github.com/Lin-Jiong-HDU/geo-optimizer/pkg/{optimizer,models,llm,config,analyzer}`

## Integration Pattern

Typical usage by backend services:

```go
import (
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
)

// 1. Initialize LLM client with provider config
llmClient := llm.NewClient(llm.Config{Provider: llm.ProviderOpenAI, ...})

// 2. Create optimizer with client
opt := optimizer.New(llmClient)

// 3. Build request with enterprise context
req := &models.OptimizationRequest{
    Content: "...",
    Enterprise: models.EnterpriseInfo{...},
    Competitors: []models.CompetitorInfo{...},
    AIPreferences: map[string]models.AIPreference{...},
}

// 4. Execute optimization
resp, err := opt.Optimize(ctx, req)
```

## Design Principles

From `docs/产品原型.md`:

- **Enterprise Isolation**: Complete data, strategy, and content separation between clients
- **Agent-Driven**: Modular collaboration between analysis/planning/writing modules
- **Evidence Traceability**: Search sources, citation chains, timestamps for reproducibility
- **Experience Solidification**: Continuous improvement of preference models and templates
- **Human-in-the-Loop**: Support both manual intervention and fully automated execution

## Some rules:

- 在实现llm客户端时，不要使用第三方的Go SDK，直接调用HTTP API
- 将代码提交到GitHub仓库前，确保使用`go fmt ./...`进行代码格式化
- 遇到不清楚的点时，可以查阅docs/下的文档

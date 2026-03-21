<div align="center">
  <img src="./geo-optimizer_logo.png" alt="geo-optimizer" width="500">
  <h1>GEO Optimizer: Make Your Content Visible to AI Engines</h1>
  <p>
    <a href="https://pkg.go.dev/github.com/Lin-Jiong-HDU/geo-optimizer"><img src="https://pkg.go.dev/badge/github.com/Lin-Jiong-HDU/geo-optimizer.svg" alt="Go Reference"></a>
    <a href="https://goreportcard.com/report/github.com/Lin-Jiong-HDU/geo-optimizer"><img src="https://goreportcard.com/badge/github.com/Lin-Jiong-HDU/geo-optimizer" alt="Go Report Card"></a>
    <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
  </p>
</div>

A **pluggable framework** for GEO (Generative Engine Optimization) in Go. Built-in strategies for common use cases, with full support for custom strategy registration.

**Why this framework?**
- 🧩 **Pluggable Architecture** - Register your own optimization strategies or use built-in ones
- 📦 **5 Built-in Strategies** - Structure, Schema, AnswerFirst, Authority, FAQ ready to use
- 🔌 **Easy Extension** - Implement the `Strategy` interface to add custom logic
- 🎯 **Composable** - Mix and match strategies for different optimization needs

## Features

- **Multi-Strategy Optimization**: 5 built-in optimization strategies that can be combined
- **LLM Abstraction**: Clean interface supporting multiple LLM providers (GLM, extensible to others)
- **Content Scoring**: Rule-based GEO quality scoring system
- **AI Platform Presets**: Pre-configured preferences for ChatGPT, Perplexity, Google AI, and Claude
- **Schema Markup**: Automatic JSON-LD structured data generation
- **Enterprise Ready**: Multi-tenant architecture with enterprise isolation

## Installation

```bash
go get github.com/Lin-Jiong-HDU/geo-optimizer
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
)

func main() {
    // 1. Create LLM client
    client, err := llm.NewClient(llm.Config{
        Provider: llm.ProviderGLM,
        APIKey:   "your-api-key",
        Model:    "glm-4.7",
    })
    if err != nil {
        log.Fatal(err)
    }

    // 2. Create optimizer
    opt := optimizer.New(client)

    // 3. Build request
    req := &models.OptimizationRequest{
        Content: "Your content here...",
        Title:   "Content Title",
        Enterprise: models.EnterpriseInfo{
            CompanyName: "Your Company",
            ProductName: "Your Product",
        },
        Strategies: []models.StrategyType{
            models.StrategyStructure,
            models.StrategySchema,
        },
    }

    // 4. Execute optimization
    resp, err := opt.Optimize(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    // 5. Use results
    fmt.Printf("Score: %.2f -> %.2f\n", resp.ScoreBefore, resp.ScoreAfter)
    fmt.Println(resp.OptimizedContent)
}
```

## Optimization Strategies

### Built-in Strategies

| Strategy | Description |
|----------|-------------|
| `StrategyStructure` | Adds clear heading hierarchy, bullet points, and organized sections |
| `StrategySchema` | Generates JSON-LD structured data markup |
| `StrategyAnswerFirst` | Moves key conclusions to the beginning |
| `StrategyAuthority` | Enhances with citations, sources, and credentials |
| `StrategyFAQ` | Generates FAQ sections for common queries |

### Register Custom Strategies

The framework is designed for extensibility. Implement the `Strategy` interface to add your own optimization logic:

```go
import strategiespkg "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer/strategies"

// Define your custom strategy
type SEOStrategy struct {
    *strategiespkg.BaseStrategy
    keywords []string
}

func NewSEOStrategy(keywords []string) *SEOStrategy {
    return &SEOStrategy{
        BaseStrategy: strategiespkg.NewBaseStrategy("seo", "SEO Optimization"),
        keywords:     keywords,
    }
}

// Implement the Strategy interface
func (s *SEOStrategy) Validate(req *models.OptimizationRequest) bool {
    return len(s.keywords) > 0
}

func (s *SEOStrategy) BuildPrompt(req *models.OptimizationRequest) string {
    return fmt.Sprintf("Optimize for keywords: %v\n\n%s", s.keywords, req.Content)
}

// Register and use your strategy
func main() {
    opt := optimizer.New(client)

    // Register custom strategy
    opt.RegisterStrategy(NewSEOStrategy([]string{"cloud", "AI", "optimization"}))

    // Use it alongside built-in strategies
    req := &models.OptimizationRequest{
        Content:    "...",
        Strategies: []models.StrategyType{"seo", models.StrategyStructure},
    }

    resp, _ := opt.Optimize(ctx, req)
}
```

### Strategy Interface

```go
type Strategy interface {
    Name() string
    Type() models.StrategyType
    Validate(req *models.OptimizationRequest) bool
    Preprocess(content string, req *models.OptimizationRequest) string
    Postprocess(content string, req *models.OptimizationRequest) string
    BuildPrompt(req *models.OptimizationRequest) string
}
```

## Architecture

```
pkg/
├── optimizer/         # Core optimization engine
│   └── strategies/    # Individual strategy implementations
├── llm/               # LLM abstraction layer
│   └── prompts/       # Prompt templates and builders
├── models/            # Data models (Request/Response)
├── analyzer/          # Content analysis and scoring
└── config/            # Configuration and AI platform presets
```

## Documentation

See [API.md](./API.md) for detailed API documentation.

## Requirements

- Go 1.21+
- LLM API access (currently supports GLM)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

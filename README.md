# GEO Optimizer

[![Go Reference](https://pkg.go.dev/badge/github.com/Lin-Jiong-HDU/geo-optimizer.svg)](https://pkg.go.dev/github.com/Lin-Jiong-HDU/geo-optimizer)
[![Go Report Card](https://goreportcard.com/badge/github.com/Lin-Jiong-HDU/geo-optimizer)](https://goreportcard.com/report/github.com/Lin-Jiong-HDU/geo-optimizer)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go library for GEO (Generative Engine Optimization) - optimizing content to improve visibility and citation rates in AI search engines like ChatGPT, Perplexity, and Google AI.

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

| Strategy | Description |
|----------|-------------|
| `StrategyStructure` | Adds clear heading hierarchy, bullet points, and organized sections |
| `StrategySchema` | Generates JSON-LD structured data markup |
| `StrategyAnswerFirst` | Moves key conclusions to the beginning |
| `StrategyAuthority` | Enhances with citations, sources, and credentials |
| `StrategyFAQ` | Generates FAQ sections for common queries |

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

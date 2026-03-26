# E2E Test Project Design

**Date**: 2026-03-26
**Status**: Approved

## Overview

Create an end-to-end test project for the geo-optimizer library. The project will be a CLI tool that validates core functionality and complete business workflows, outputting results to both console and Markdown reports.

## Requirements

| Item | Choice |
|------|--------|
| Test Scenarios | Basic functionality + Full business flow |
| Output Format | CLI tool |
| LLM Configuration | .env file |
| Project Location | examples/ directory |
| Result Output | Markdown + Console |

## Project Structure

```
examples/e2e/
├── main.go              # CLI entry point, command parsing
├── go.mod               # Independent module, references main repo
├── .env.example         # Configuration example
├── scenarios/
│   ├── scenario.go      # Scenario interface definition
│   ├── basic.go         # Basic functionality tests
│   └── full_flow.go     # Full business flow tests
├── config/
│   └── config.go        # .env loading, config structure
├── reporter/
│   └── reporter.go      # Console + Markdown reporter
└── output/              # Test report output (gitignore)
```

## Scenario Interface

```go
// scenarios/scenario.go
type Result struct {
    Name        string
    Success     bool
    Duration    time.Duration
    Input       string
    Output      string
    ScoreBefore float64
    ScoreAfter  float64
    Error       string
}

type Scenario interface {
    Name() string
    Description() string
    Run(ctx context.Context, client llm.LLMClient, opt *optimizer.Optimizer) (*Result, error)
}
```

## Test Scenarios

### Basic Functionality (basic.go)

- `TestChat`: Validate LLM client Chat interface
- `TestScore`: Validate Scorer scoring functionality
- `TestOptimizeWithStrategy`: Validate single strategy optimization

### Full Business Flow (full_flow.go)

- Use realistic enterprise and product content
- Call complete `Optimize()` pipeline
- Validate all response fields (optimized content, Schema, FAQ, score comparison, etc.)

## CLI Commands

```bash
# Run all scenarios
go run .

# Run specific scenario
go run . --scenario=basic
go run . --scenario=full_flow

# Specify output directory
go run . --output=./reports

# Show help
go run . --help
```

**Parameters**:
- `--scenario`: Specify scenario, default `all`
- `--output`: Report output directory, default `./output`
- `--verbose`: Detailed log output

## Report Format

### Console Output

```
[E2E Test] Starting...
├── Loading config from .env ✓
├── Initializing LLM client ✓
│
├── [1/2] Basic Functionality
│   ├── TestChat .................. ✓ (0.8s)
│   ├── TestScore ................. ✓ (0.2s)
│   └── TestOptimizeWithStrategy .. ✓ (2.1s)
│
├── [2/2] Full Business Flow
│   └── TestOptimize .............. ✓ (5.3s)
│       Score: 45.2 → 78.6 (+33.4)
│
└── All tests passed! Report: output/e2e-report-2026-03-26.md
```

### Markdown Report

```markdown
# E2E Test Report
**Date**: 2026-03-26 14:30:00
**Total Duration**: 8.4s
**Result**: ✓ All Passed

## Scenario: Basic Functionality
### TestChat
- **Input**: 什么是AI？
- **Output**: AI是人工智能...
- **Duration**: 0.8s

### TestScore
- **Score Before**: 32.5
- **Score After**: N/A (single test)

## Scenario: Full Business Flow
### TestOptimize
- **Input**: [Enterprise content...]
- **Optimized Content**: [Optimized content...]
- **Score**: 45.2 → 78.6 (+33.4)
- **Schema Markup**: ✓ Generated
- **FAQ Section**: ✓ Generated
```

## Configuration

### .env.example

```
GLM_API_KEY=your_api_key_here
GLM_MODEL=glm-4-flash
GLM_BASE_URL=   # Optional, custom endpoint
GLM_TIMEOUT=300
```

### config/config.go

```go
type Config struct {
    APIKey  string
    Model   string  // Default: glm-4-flash
    BaseURL string  // Optional
    Timeout int     // Default: 300s
}

func Load() (*Config, error) {
    // Load from .env file
    // Support lookup from project root or examples/e2e directory
}
```

## Dependencies

### go.mod

```
module github.com/Lin-Jiong-HDU/geo-optimizer/examples/e2e

go 1.25.5

require github.com/Lin-Jiong-HDU/geo-optimizer v0.0.0

replace github.com/Lin-Jiong-HDU/geo-optimizer => ../..
```

## Module Responsibilities

| Module | Responsibility |
|--------|----------------|
| `main.go` | CLI parsing, scenario orchestration, report aggregation |
| `config/` | .env loading, configuration management |
| `scenarios/` | Test scenario implementation, unified interface |
| `reporter/` | Console progress + Markdown report |

## Data Flow

1. CLI parses arguments → Load .env config
2. Initialize LLMClient and Optimizer
3. Execute scenarios sequentially, collect Results
4. Output console progress in real-time
5. Generate Markdown report to output/

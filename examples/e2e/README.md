# E2E Test Project

End-to-end testing for the geo-optimizer library.

## Overview

This project provides CLI-based E2E tests that validate the geo-optimizer's core functionality and complete business workflows. Tests connect to real LLM APIs and generate both console output and Markdown reports.

## Prerequisites

- Go 1.25.5+
- GLM API key (from Zhipu AI)

## Setup

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit `.env` and add your GLM API key:
```
GLM_API_KEY=your_actual_api_key_here
```

## Usage

```bash
# Run all scenarios
go run .

# Run specific scenario
go run . --scenario=basic
go run . --scenario=full_flow

# Verbose output
go run . --verbose

# Custom output directory
go run . --output=./reports

# Build binary
go build -o e2e
./e2e --scenario=all
```

## Scenarios

### Basic Functionality
Validates core APIs:
- `TestChat`: LLM client communication
- `TestScore`: Content scoring
- `TestOptimizeWithStrategy`: Single strategy optimization

### Full Business Flow
Tests the complete `Optimize()` pipeline with realistic enterprise content, validating:
- Optimized content generation
- Schema markup
- FAQ section
- Token tracking

## Output

Reports are generated in the `output/` directory with timestamps:
```
output/e2e-report-2026-03-26-14-30-00.md
```

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--scenario` | Scenario to run: `all`, `basic`, `full_flow` | `all` |
| `--output` | Output directory for reports | `./output` |
| `--verbose` | Enable detailed error messages | `false` |

## Exit Codes

- `0`: All tests passed
- `1`: One or more tests failed

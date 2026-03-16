# Contributing to GEO Optimizer

Thank you for considering contributing to GEO Optimizer!

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Code Guidelines](#code-guidelines)
- [Commit Convention](#commit-convention)
- [Pull Request Process](#pull-request-process)

## Code of Conduct

- Be respectful to all contributors
- Keep discussions professional and constructive
- Accept constructive criticism gracefully

## How to Contribute

### Reporting Bugs

If you find a bug, please submit an issue with:

1. **Description** - Clear description of what happened
2. **Steps to Reproduce** - How to trigger the bug
3. **Expected Behavior** - What you expected to happen
4. **Actual Behavior** - What actually happened
5. **Environment** - Go version, OS, etc.

### Suggesting Features

For new features, please submit an issue describing:

1. **Use Case** - What problem does this solve
2. **Proposed Solution** - How you think it should work
3. **Alternatives** - Other approaches you've considered

### Submitting Code

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.21+
- Git
- LLM API Key (for integration testing)

### Clone and Build

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/geo-optimizer.git
cd geo-optimizer

# Download dependencies
go mod download

# Run tests
go test ./...

# Build
go build ./...
```

### Running Tests

```bash
# Run all tests
go test ./... -v

# Run specific package tests
go test ./pkg/optimizer -v

# Run single test
go test ./pkg/analyzer -run TestScorer -v
```

### Code Formatting

```bash
# Format code before committing
go fmt ./...
```

## Code Guidelines

### Project Structure

```
pkg/
├── optimizer/         # Core optimization engine
│   └── strategies/    # Strategy implementations
├── llm/               # LLM abstraction layer
│   └── prompts/       # Prompt templates and builders
├── models/            # Data models (Request/Response)
├── analyzer/          # Content analysis and scoring
└── config/            # Configuration and presets
```

### Adding a New Strategy

Implement the `Strategy` interface:

```go
import (
    "github.com/geo-team-red/geo-optimizer/pkg/models"
    "github.com/geo-team-red/geo-optimizer/pkg/optimizer/strategies"
    "github.com/geo-team-red/geo-optimizer/pkg/llm/prompts"
)

type MyStrategy struct {
    *strategies.BaseStrategy
}

func NewMyStrategy() *MyStrategy {
    return &MyStrategy{
        BaseStrategy: strategies.NewBaseStrategy("my_strategy", "My Strategy"),
    }
}

func (s *MyStrategy) Validate(req *models.OptimizationRequest) bool {
    return req.Content != ""
}

func (s *MyStrategy) BuildPrompt(req *models.OptimizationRequest) string {
    return prompts.NewBuilder().BuildStrategyPrompt(models.StrategyType("my_strategy"), req)
}
```

### Adding a New LLM Provider

Implement the `LLMClient` interface:

```go
type LLMClient interface {
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}
```

### Code Style

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Format code with `gofmt`
- Add comments for exported functions and types
- Error messages should not be capitalized or end with punctuation

## Commit Convention

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation |
| `style` | Code formatting (no functionality change) |
| `refactor` | Code refactoring |
| `test` | Testing |
| `chore` | Build, dependencies, etc. |

### Examples

```
feat(optimizer): add keyword density strategy

- Implement KeywordDensityStrategy
- Add configuration for density threshold
- Add unit tests

Closes #123
```

## Pull Request Process

1. **Ensure tests pass** - `go test ./...`
2. **Format code** - `go fmt ./...`
3. **Update documentation** - Update `API.md` for API changes
4. **Add tests** - New features need test coverage
5. **One PR per feature** - Keep PRs focused

### PR Title Format

```
<type>(<scope>): <description>
```

Example: `feat(strategies): add keyword optimization strategy`

### Checklist

- [ ] Code passes all tests
- [ ] Code is formatted with `gofmt`
- [ ] New code has appropriate comments
- [ ] Documentation is updated
- [ ] No new warnings introduced

## License

By submitting code, you agree that your contributions will be licensed under the [MIT License](LICENSE).

---

Questions? Feel free to ask in [Issues](https://github.com/geo-team-red/geo-optimizer/issues)!

# E2E Test Project Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a CLI-based e2e test project that validates geo-optimizer's core functionality and complete business workflows.

**Architecture:** Independent Go module under examples/e2e/ with scenarios implementing a unified interface, config loaded from .env, and dual output (console + Markdown reports).

**Tech Stack:** Go 1.25.5, godotenv (for .env parsing), geo-optimizer library

---

## File Structure

| File | Purpose |
|------|---------|
| `examples/e2e/go.mod` | Module definition with replace directive |
| `examples/e2e/main.go` | CLI entry point, orchestrates scenarios |
| `examples/e2e/.env.example` | Configuration template |
| `examples/e2e/config/config.go` | .env loading and Config struct |
| `examples/e2e/scenarios/scenario.go` | Scenario interface and Result struct |
| `examples/e2e/scenarios/basic.go` | Basic functionality tests |
| `examples/e2e/scenarios/full_flow.go` | Full business flow tests |
| `examples/e2e/reporter/reporter.go` | Console and Markdown output |

---

## Chunk 1: Project Setup and Config

### Task 1: Create Project Structure

**Files:**
- Create: `examples/e2e/go.mod`
- Create: `examples/e2e/.env.example`

- [ ] **Step 1: Create examples/e2e directory**

```bash
mkdir -p examples/e2e/config examples/e2e/scenarios examples/e2e/reporter examples/e2e/output
```

- [ ] **Step 2: Create go.mod with replace directive**

```go
module github.com/Lin-Jiong-HDU/geo-optimizer/examples/e2e

go 1.25.5

require github.com/Lin-Jiong-HDU/geo-optimizer v0.0.0

replace github.com/Lin-Jiong-HDU/geo-optimizer => ../..
```

- [ ] **Step 3: Create .env.example**

```
GLM_API_KEY=your_api_key_here
GLM_MODEL=glm-4-flash
GLM_BASE_URL=
GLM_TIMEOUT=300
```

- [ ] **Step 4: Create .gitignore for output directory**

Create `examples/e2e/output/.gitignore`:
```
*
!.gitignore
```

- [ ] **Step 5: Commit project structure**

```bash
git add examples/e2e/
git commit -m "feat(e2e): create project structure and configuration"
```

---

### Task 2: Implement Config Module

**Files:**
- Create: `examples/e2e/config/config.go`

- [ ] **Step 1: Write config.go with .env loading**

```go
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the e2e test configuration.
type Config struct {
	APIKey  string
	Model   string
	BaseURL string
	Timeout int
}

// Load loads configuration from .env file.
func Load() (*Config, error) {
	cfg := &Config{
		Model:   "glm-4-flash",
		Timeout: 300,
	}

	// Try multiple paths to find .env
	paths := []string{
		".env",
		"../.env",
		"../../.env",
	}

	var envPath string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			envPath = p
			break
		}
	}

	if envPath == "" {
		return nil, fmt.Errorf(".env file not found, please copy .env.example to .env and configure")
	}

	if err := loadEnvFile(envPath); err != nil {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	cfg.APIKey = os.Getenv("GLM_API_KEY")
	if model := os.Getenv("GLM_MODEL"); model != "" {
		cfg.Model = model
	}
	if baseURL := os.Getenv("GLM_BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if timeout := os.Getenv("GLM_TIMEOUT"); timeout != "" {
		fmt.Sscanf(timeout, "%d", &cfg.Timeout)
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("GLM_API_KEY is required in .env file")
	}

	return cfg, nil
}

// loadEnvFile loads environment variables from a file.
func loadEnvFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
```

- [ ] **Step 2: Verify config compiles**

```bash
cd examples/e2e && go build ./config/
```

Expected: No errors

- [ ] **Step 3: Commit config module**

```bash
git add examples/e2e/config/
git commit -m "feat(e2e): add config module with .env loading"
```

---

## Chunk 2: Scenario Interface and Reporter

### Task 3: Implement Scenario Interface

**Files:**
- Create: `examples/e2e/scenarios/scenario.go`

- [ ] **Step 1: Write scenario interface and Result struct**

```go
package scenarios

import (
	"context"
	"time"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
)

// Result represents the result of a single test within a scenario.
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

// ScenarioResult represents the result of an entire scenario.
type ScenarioResult struct {
	ScenarioName string
	Description  string
	Results      []Result
	TotalDuration time.Duration
	AllPassed    bool
}

// Scenario defines the interface for test scenarios.
type Scenario interface {
	Name() string
	Description() string
	Run(ctx context.Context, client llm.LLMClient, opt *optimizer.Optimizer) (*ScenarioResult, error)
}

// NewResult creates a new Result with default values.
func NewResult(name string) *Result {
	return &Result{
		Name:    name,
		Success: false,
	}
}

// SetSuccess marks the result as successful with output.
func (r *Result) SetSuccess(output string, duration time.Duration) {
	r.Success = true
	r.Output = output
	r.Duration = duration
}

// SetFailure marks the result as failed with error message.
func (r *Result) SetFailure(errMsg string, duration time.Duration) {
	r.Success = false
	r.Error = errMsg
	r.Duration = duration
}
```

- [ ] **Step 2: Verify scenario compiles**

```bash
cd examples/e2e && go build ./scenarios/
```

Expected: No errors

- [ ] **Step 3: Commit scenario interface**

```bash
git add examples/e2e/scenarios/scenario.go
git commit -m "feat(e2e): add Scenario interface and Result types"
```

---

### Task 4: Implement Reporter

**Files:**
- Create: `examples/e2e/reporter/reporter.go`

- [ ] **Step 1: Write reporter with console and Markdown output**

```go
package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lin-Jiong-HDU/geo-optimizer/examples/e2e/scenarios"
)

// Reporter handles console output and Markdown report generation.
type Reporter struct {
	verbose     bool
	outputDir   string
	results     []scenarios.ScenarioResult
	startTime   time.Time
}

// New creates a new Reporter.
func New(outputDir string, verbose bool) *Reporter {
	return &Reporter{
		verbose:   verbose,
		outputDir: outputDir,
		results:   make([]scenarios.ScenarioResult, 0),
	}
}

// Start prints the test start message.
func (r *Reporter) Start() {
	r.startTime = time.Now()
	fmt.Println("[E2E Test] Starting...")
}

// PrintConfig prints config loading status.
func (r *Reporter) PrintConfig(success bool) {
	if success {
		fmt.Println("├── Loading config from .env ✓")
	} else {
		fmt.Println("├── Loading config from .env ✗")
	}
}

// PrintInit prints initialization status.
func (r *Reporter) PrintInit(success bool) {
	if success {
		fmt.Println("├── Initializing LLM client ✓")
	} else {
		fmt.Println("├── Initializing LLM client ✗")
	}
	fmt.Println("│")
}

// PrintScenarioStart prints scenario header.
func (r *Reporter) PrintScenarioStart(index, total int, name string) {
	fmt.Printf("├── [%d/%d] %s\n", index, total, name)
}

// PrintResult prints a single test result.
func (r *Reporter) PrintResult(result scenarios.Result) {
	status := "✓"
	if !result.Success {
		status = "✗"
	}
	dots := strings.Repeat(".", 40-len(result.Name))
	fmt.Printf("│   ├── %s%s %s (%.1fs)\n", result.Name, dots, status, result.Duration.Seconds())

	if !result.Success && r.verbose {
		fmt.Printf("│       Error: %s\n", result.Error)
	}

	if result.ScoreBefore > 0 && result.ScoreAfter > 0 {
		diff := result.ScoreAfter - result.ScoreBefore
		fmt.Printf("│       Score: %.1f → %.1f (%+.1f)\n", result.ScoreBefore, result.ScoreAfter, diff)
	}
}

// AddResult stores scenario result for report generation.
func (r *Reporter) AddResult(result scenarios.ScenarioResult) {
	r.results = append(r.results, result)
}

// Finish prints final summary and generates Markdown report.
func (r *Reporter) Finish() (string, error) {
	totalDuration := time.Since(r.startTime)
	allPassed := true
	for _, sr := range r.results {
		if !sr.AllPassed {
			allPassed = false
			break
		}
	}

	status := "✓ All Passed"
	if !allPassed {
		status = "✗ Some Failed"
	}

	reportPath, err := r.generateMarkdownReport(totalDuration, allPassed)
	if err != nil {
		fmt.Printf("└── Report generation failed: %v\n", err)
		return "", err
	}

	fmt.Printf("└── %s! Report: %s\n", status, reportPath)
	return reportPath, nil
}

// generateMarkdownReport creates the Markdown report file.
func (r *Reporter) generateMarkdownReport(totalDuration time.Duration, allPassed bool) (string, error) {
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return "", err
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("e2e-report-%s.md", timestamp)
	path := filepath.Join(r.outputDir, filename)

	var sb strings.Builder

	status := "✓ All Passed"
	if !allPassed {
		status = "✗ Some Failed"
	}

	sb.WriteString("# E2E Test Report\n\n")
	sb.WriteString(fmt.Sprintf("**Date**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Total Duration**: %.1fs\n", totalDuration.Seconds()))
	sb.WriteString(fmt.Sprintf("**Result**: %s\n\n", status))

	for _, sr := range r.results {
		sb.WriteString(fmt.Sprintf("## Scenario: %s\n", sr.ScenarioName))
		sb.WriteString(fmt.Sprintf("*%s*\n\n", sr.Description))

		for _, res := range sr.Results {
			sb.WriteString(fmt.Sprintf("### %s\n", res.Name))
			sb.WriteString(fmt.Sprintf("- **Status**: %s\n", boolToStr(res.Success)))
			sb.WriteString(fmt.Sprintf("- **Duration**: %.1fs\n", res.Duration.Seconds()))

			if res.Input != "" {
				sb.WriteString(fmt.Sprintf("- **Input**: %s\n", truncate(res.Input, 200)))
			}

			if res.Output != "" {
				sb.WriteString(fmt.Sprintf("- **Output**: %s\n", truncate(res.Output, 500)))
			}

			if res.Error != "" {
				sb.WriteString(fmt.Sprintf("- **Error**: %s\n", res.Error))
			}

			if res.ScoreBefore > 0 || res.ScoreAfter > 0 {
				sb.WriteString(fmt.Sprintf("- **Score**: %.1f → %.1f\n", res.ScoreBefore, res.ScoreAfter))
			}

			sb.WriteString("\n")
		}
	}

	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		return "", err
	}

	return path, nil
}

func boolToStr(b bool) string {
	if b {
		return "✓ Passed"
	}
	return "✗ Failed"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
```

- [ ] **Step 2: Verify reporter compiles**

```bash
cd examples/e2e && go build ./reporter/
```

Expected: No errors

- [ ] **Step 3: Commit reporter**

```bash
git add examples/e2e/reporter/
git commit -m "feat(e2e): add reporter with console and Markdown output"
```

---

## Chunk 3: Test Scenarios

### Task 5: Implement Basic Scenario

**Files:**
- Create: `examples/e2e/scenarios/basic.go`

- [ ] **Step 1: Write basic scenario with three tests**

```go
package scenarios

import (
	"context"
	"fmt"
	"time"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
)

// BasicScenario tests basic functionality of the geo-optimizer.
type BasicScenario struct{}

// NewBasicScenario creates a new BasicScenario.
func NewBasicScenario() *BasicScenario {
	return &BasicScenario{}
}

// Name returns the scenario name.
func (s *BasicScenario) Name() string {
	return "Basic Functionality"
}

// Description returns the scenario description.
func (s *BasicScenario) Description() string {
	return "Validates core APIs: Chat, Score, and OptimizeWithStrategy"
}

// Run executes the basic functionality tests.
func (s *BasicScenario) Run(ctx context.Context, client llm.LLMClient, opt *optimizer.Optimizer) (*ScenarioResult, error) {
	result := &ScenarioResult{
		ScenarioName: s.Name(),
		Description:  s.Description(),
		Results:      make([]Result, 0),
		AllPassed:    true,
	}

	startTime := time.Now()

	// Test 1: Chat
	chatResult := s.testChat(ctx, client)
	result.Results = append(result.Results, chatResult)
	if !chatResult.Success {
		result.AllPassed = false
	}

	// Test 2: Score
	scoreResult := s.testScore(ctx, opt)
	result.Results = append(result.Results, scoreResult)
	if !scoreResult.Success {
		result.AllPassed = false
	}

	// Test 3: OptimizeWithStrategy
	optResult := s.testOptimizeWithStrategy(ctx, opt)
	result.Results = append(result.Results, optResult)
	if !optResult.Success {
		result.AllPassed = false
	}

	result.TotalDuration = time.Since(startTime)
	return result, nil
}

func (s *BasicScenario) testChat(ctx context.Context, client llm.LLMClient) Result {
	result := NewResult("TestChat")
	start := time.Now()

	input := "什么是AI？"
	resp, err := client.Chat(ctx, &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: input},
		},
	})

	if err != nil {
		result.SetFailure(err.Error(), time.Since(start))
		return *result
	}

	result.Input = input
	result.SetSuccess(resp.Content, time.Since(start))
	return *result
}

func (s *BasicScenario) testScore(ctx context.Context, opt *optimizer.Optimizer) Result {
	result := NewResult("TestScore")
	start := time.Now()

	content := "人工智能正在改变世界。它可以帮助我们解决很多问题。"
	score, err := opt.Scorer().Score(ctx, content)

	if err != nil {
		result.SetFailure(err.Error(), time.Since(start))
		return *result
	}

	result.Input = content
	result.ScoreBefore = score.OverallScore()
	result.SetSuccess(fmt.Sprintf("Score: %.1f", score.OverallScore()), time.Since(start))
	return *result
}

func (s *BasicScenario) testOptimizeWithStrategy(ctx context.Context, opt *optimizer.Optimizer) Result {
	result := NewResult("TestOptimizeWithStrategy")
	start := time.Now()

	req := &models.OptimizationRequest{
		Title:   "AI技术简介",
		Content: "AI是人工智能的缩写，是计算机科学的一个分支。",
		Enterprise: models.EnterpriseInfo{
			CompanyName:        "测试公司",
			ProductName:        "测试产品",
			ProductDescription: "这是一个测试产品",
		},
		Strategies: []models.StrategyType{models.StrategyStructure},
	}

	resp, err := opt.OptimizeWithStrategy(ctx, req, models.StrategyStructure)
	if err != nil {
		result.SetFailure(err.Error(), time.Since(start))
		return *result
	}

	result.Input = req.Content
	result.Output = resp.OptimizedContent
	result.ScoreBefore = resp.ScoreBefore
	result.ScoreAfter = resp.ScoreAfter
	result.SetSuccess(fmt.Sprintf("Tokens used: %d", resp.TokensUsed), time.Since(start))
	return *result
}
```

- [ ] **Step 2: Verify basic scenario compiles**

```bash
cd examples/e2e && go build ./scenarios/
```

Expected: No errors

- [ ] **Step 3: Commit basic scenario**

```bash
git add examples/e2e/scenarios/basic.go
git commit -m "feat(e2e): add basic functionality scenario"
```

---

### Task 6: Implement Full Flow Scenario

**Files:**
- Create: `examples/e2e/scenarios/full_flow.go`

- [ ] **Step 1: Write full flow scenario with realistic data**

```go
package scenarios

import (
	"context"
	"fmt"
	"time"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
)

// FullFlowScenario tests the complete optimization workflow.
type FullFlowScenario struct{}

// NewFullFlowScenario creates a new FullFlowScenario.
func NewFullFlowScenario() *FullFlowScenario {
	return &FullFlowScenario{}
}

// Name returns the scenario name.
func (s *FullFlowScenario) Name() string {
	return "Full Business Flow"
}

// Description returns the scenario description.
func (s *FullFlowScenario) Description() string {
	return "Tests complete Optimize() pipeline with realistic enterprise content"
}

// Run executes the full business flow test.
func (s *FullFlowScenario) Run(ctx context.Context, client llm.LLMClient, opt *optimizer.Optimizer) (*ScenarioResult, error) {
	result := &ScenarioResult{
		ScenarioName: s.Name(),
		Description:  s.Description(),
		Results:      make([]Result, 0),
		AllPassed:    true,
	}

	startTime := time.Now()

	// Run the full optimization test
	testResult := s.testFullOptimize(ctx, opt)
	result.Results = append(result.Results, testResult)
	if !testResult.Success {
		result.AllPassed = false
	}

	result.TotalDuration = time.Since(startTime)
	return result, nil
}

func (s *FullFlowScenario) testFullOptimize(ctx context.Context, opt *optimizer.Optimizer) Result {
	result := NewResult("TestOptimize")
	start := time.Now()

	// Use realistic enterprise content
	req := &models.OptimizationRequest{
		Title: "企业数字化转型解决方案",
		Content: `企业数字化转型是当今商业环境中的重要议题。

随着技术的快速发展，企业需要适应新的市场环境。数字化转型可以帮助企业提高效率、降低成本、增强竞争力。

我们的解决方案包括：
- 云计算基础设施
- 数据分析平台
- 人工智能应用

这些技术可以帮助企业实现业务目标。`,
		Enterprise: models.EnterpriseInfo{
			CompanyName:        "智云科技",
			CompanyWebsite:     "https://example.com",
			CompanyDescription: "专注于企业数字化转型解决方案",
			ProductName:        "智云企业平台",
			ProductURL:         "https://example.com/platform",
			ProductDescription: "一站式企业数字化解决方案",
			ProductFeatures: []string{
				"云端部署",
				"数据可视化",
				"AI智能分析",
				"多端协同",
			},
			USP: []string{
				"行业领先的AI技术",
				"99.9%服务可用性",
				"7x24小时技术支持",
			},
			BrandVoice:       "专业、可靠、创新",
			TargetAudience:   "中大型企业决策者",
			ValueProposition: "帮助企业快速实现数字化转型",
		},
		Competitors: []models.CompetitorInfo{
			{
				Name:       "竞争对手A",
				Website:    "https://competitor-a.com",
				Weaknesses: []string{"功能单一", "缺乏AI支持"},
			},
		},
		TargetAI:   []string{"chatgpt", "perplexity"},
		Keywords:   []string{"数字化转型", "企业解决方案", "云计算", "AI"},
		Strategies: []models.StrategyType{models.StrategyStructure, models.StrategySchema, models.StrategyFAQ},
		Tone:       "professional",
		Industry:   "technology",
	}

	resp, err := opt.Optimize(ctx, req)
	if err != nil {
		result.SetFailure(err.Error(), time.Since(start))
		return *result
	}

	// Validate response fields
	var validationErrors []string
	if resp.OptimizedContent == "" {
		validationErrors = append(validationErrors, "OptimizedContent is empty")
	}
	if resp.SchemaMarkup == "" {
		validationErrors = append(validationErrors, "SchemaMarkup is empty")
	}
	if resp.FAQSection == "" {
		validationErrors = append(validationErrors, "FAQSection is empty")
	}
	if resp.TokensUsed == 0 {
		validationErrors = append(validationErrors, "TokensUsed is 0")
	}

	if len(validationErrors) > 0 {
		result.SetFailure(fmt.Sprintf("Validation failed: %v", validationErrors), time.Since(start))
		return *result
	}

	result.Input = req.Content
	result.Output = fmt.Sprintf("Content length: %d, Schema: %t, FAQ: %t",
		len(resp.OptimizedContent), resp.SchemaMarkup != "", resp.FAQSection != "")
	result.ScoreBefore = resp.ScoreBefore
	result.ScoreAfter = resp.ScoreAfter
	result.SetSuccess(fmt.Sprintf("All fields validated, Tokens: %d", resp.TokensUsed), time.Since(start))
	return *result
}
```

- [ ] **Step 2: Verify full flow scenario compiles**

```bash
cd examples/e2e && go build ./scenarios/
```

Expected: No errors

- [ ] **Step 3: Commit full flow scenario**

```bash
git add examples/e2e/scenarios/full_flow.go
git commit -m "feat(e2e): add full business flow scenario"
```

---

## Chunk 4: Main Entry Point

### Task 7: Implement Main CLI

**Files:**
- Create: `examples/e2e/main.go`

- [ ] **Step 1: Write main.go with CLI parsing and orchestration**

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Lin-Jiong-HDU/geo-optimizer/examples/e2e/config"
	"github.com/Lin-Jiong-HDU/geo-optimizer/examples/e2e/reporter"
	"github.com/Lin-Jiong-HDU/geo-optimizer/examples/e2e/scenarios"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
)

func main() {
	// Parse command line flags
	scenarioFlag := flag.String("scenario", "all", "Scenario to run: all, basic, full_flow")
	outputFlag := flag.String("output", "./output", "Output directory for reports")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose output")
	flag.Parse()

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nInterrupted, cleaning up...")
		cancel()
	}()

	// Initialize reporter
	rep := reporter.New(*outputFlag, *verboseFlag)
	rep.Start()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		rep.PrintConfig(false)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	rep.PrintConfig(true)

	// Initialize LLM client
	llmClient, err := llm.NewClient(llm.Config{
		Provider:    llm.ProviderGLM,
		APIKey:      cfg.APIKey,
		Model:       cfg.Model,
		BaseURL:     cfg.BaseURL,
		MaxTokens:   65536,
		Temperature: 0.7,
		Timeout:     cfg.Timeout,
	})
	if err != nil {
		rep.PrintInit(false)
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	rep.PrintInit(true)

	// Initialize optimizer
	opt := optimizer.New(llmClient)

	// Build scenario list
	var scenarioList []scenarios.Scenario
	switch *scenarioFlag {
	case "basic":
		scenarioList = []scenarios.Scenario{scenarios.NewBasicScenario()}
	case "full_flow":
		scenarioList = []scenarios.Scenario{scenarios.NewFullFlowScenario()}
	default:
		scenarioList = []scenarios.Scenario{
			scenarios.NewBasicScenario(),
			scenarios.NewFullFlowScenario(),
		}
	}

	// Run scenarios
	allPassed := true
	for i, s := range scenarioList {
		rep.PrintScenarioStart(i+1, len(scenarioList), s.Name())

		result, err := s.Run(ctx, llmClient, opt)
		if err != nil {
			fmt.Printf("│   └── Scenario failed: %v\n", err)
			allPassed = false
			continue
		}

		for _, r := range result.Results {
			rep.PrintResult(r)
		}

		rep.AddResult(*result)
		if !result.AllPassed {
			allPassed = false
		}
	}

	// Generate final report
	reportPath, err := rep.Finish()
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err)
		os.Exit(1)
	}

	// Exit with appropriate code
	if !allPassed {
		os.Exit(1)
	}

	fmt.Printf("\nReport saved to: %s\n", reportPath)
}
```

- [ ] **Step 2: Verify main compiles**

```bash
cd examples/e2e && go build .
```

Expected: No errors

- [ ] **Step 3: Commit main entry point**

```bash
git add examples/e2e/main.go
git commit -m "feat(e2e): add main CLI entry point"
```

---

### Task 8: Add Scorer Accessor

**Files:**
- Modify: `pkg/optimizer/optimizer.go`

- [ ] **Step 1: Add Scorer accessor method to optimizer**

Add this method to `pkg/optimizer/optimizer.go` (after the `RegisterDefaultStrategies` function):

```go
// Scorer returns the scorer instance for direct access.
func (o *Optimizer) Scorer() *analyzer.Scorer {
	return o.scorer
}
```

- [ ] **Step 2: Verify build passes**

```bash
go build ./...
```

Expected: No errors

- [ ] **Step 3: Commit accessor**

```bash
git add pkg/optimizer/optimizer.go
git commit -m "feat(optimizer): add Scorer accessor method"
```

---

### Task 9: Final Integration Test

**Files:**
- None (verification only)

- [ ] **Step 1: Copy .env.example to .env**

```bash
cp examples/e2e/.env.example examples/e2e/.env
```

- [ ] **Step 2: Run e2e tests**

```bash
cd examples/e2e && go run . --scenario=basic
```

Expected: Tests run and produce output

- [ ] **Step 3: Format all code**

```bash
go fmt ./... && cd examples/e2e && go fmt ./...
```

- [ ] **Step 4: Final commit**

```bash
git add .
git commit -m "feat(e2e): complete e2e test project implementation"
```

---

## Summary

| Task | Description |
|------|-------------|
| 1 | Create project structure (go.mod, .env.example, directories) |
| 2 | Implement config module for .env loading |
| 3 | Implement Scenario interface and Result types |
| 4 | Implement Reporter for console and Markdown output |
| 5 | Implement Basic scenario (Chat, Score, OptimizeWithStrategy) |
| 6 | Implement Full Flow scenario (complete Optimize pipeline) |
| 7 | Implement main.go CLI entry point |
| 8 | Add Scorer accessor to optimizer |
| 9 | Final integration test and cleanup |

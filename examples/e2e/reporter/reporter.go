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
	verbose   bool
	outputDir string
	results   []scenarios.ScenarioResult
	startTime time.Time
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

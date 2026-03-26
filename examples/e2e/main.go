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

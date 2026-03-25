package strategies

import (
	"fmt"
	"strings"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// StructureStrategy optimizes content structure for AI search engine preferences.
type StructureStrategy struct {
	*BaseStrategy
}

// NewStructureStrategy creates a new structure strategy.
func NewStructureStrategy() *StructureStrategy {
	return &StructureStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategyStructure, "structure"),
	}
}

// Validate checks if the strategy is applicable.
func (s *StructureStrategy) Validate(req *models.OptimizationRequest) bool {
	return req.Content != ""
}

// Preprocess preprocesses content before optimization.
func (s *StructureStrategy) Preprocess(content string, req *models.OptimizationRequest) string {
	if s.hasGoodStructure(content) {
		return content
	}

	return fmt.Sprintf("# %s\n\n%s", req.Title, content)
}

// Postprocess postprocesses content after optimization.
func (s *StructureStrategy) Postprocess(content string, req *models.OptimizationRequest) string {
	lines := strings.Split(content, "\n")
	var cleaned []string
	emptyCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			emptyCount++
			if emptyCount <= 2 {
				cleaned = append(cleaned, line)
			}
		} else {
			emptyCount = 0
			cleaned = append(cleaned, line)
		}
	}

	return strings.Join(cleaned, "\n")
}

// BuildPrompt builds the structure optimization prompt.
func (s *StructureStrategy) BuildPrompt(req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildStrategyPrompt(models.StrategyStructure, req)
}

// BuildPromptWithContent builds the prompt with specified content.
func (s *StructureStrategy) BuildPromptWithContent(content string, req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildStrategyPromptWithContent(models.StrategyStructure, content, req)
}

// hasGoodStructure checks if content already has good structure.
func (s *StructureStrategy) hasGoodStructure(content string) bool {
	hasHeading := strings.Contains(content, "# ") || strings.Contains(content, "## ")
	hasList := strings.Contains(content, "- ") || strings.Contains(content, "* ") ||
		strings.Contains(content, "1. ") || strings.Contains(content, "2. ")
	hasParagraph := strings.Contains(content, "\n\n")
	return hasHeading && (hasList || hasParagraph)
}

// GetStructureScore returns the structure score of the content.
func (s *StructureStrategy) GetStructureScore(content string) float64 {
	score := 0.0

	headings := strings.Count(content, "#")
	if headings > 0 {
		score += 15.0
		if headings >= 3 {
			score += 10.0
		}
		if headings >= 5 {
			score += 5.0
		}
	}

	if strings.Contains(content, "- ") || strings.Contains(content, "* ") {
		score += 10.0
	}
	if strings.Contains(content, "1.") || strings.Contains(content, "2.") {
		score += 10.0
	}

	paragraphs := strings.Split(content, "\n\n")
	if len(paragraphs) > 1 {
		score += 15.0
	}
	if len(paragraphs) > 3 {
		score += 10.0
	}

	sections := strings.Count(content, "##")
	if sections >= 2 {
		score += 15.0
	}
	if sections >= 4 {
		score += 10.0
	}

	if score > 100 {
		score = 100
	}

	return score
}

package strategies

import (
	"fmt"
	"strings"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// StructureStrategy 结构化优化策略
// 优化内容的结构，使其更符合AI搜索引擎的偏好
type StructureStrategy struct {
	*BaseStrategy
}

// NewStructureStrategy 创建结构化策略
func NewStructureStrategy() *StructureStrategy {
	return &StructureStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategyStructure, "structure"),
	}
}

// Validate 验证策略是否适用
func (s *StructureStrategy) Validate(req *models.OptimizationRequest) bool {
	// 结构化策略总是适用
	return req.Content != ""
}

// Preprocess 预处理内容
func (s *StructureStrategy) Preprocess(content string, req *models.OptimizationRequest) string {
	// 检查内容是否已经有良好的结构
	if s.hasGoodStructure(content) {
		return content
	}

	// 如果结构不好，添加基础结构标记
	return fmt.Sprintf("# %s\n\n%s", req.Title, content)
}

// Postprocess 后处理内容
func (s *StructureStrategy) Postprocess(content string, req *models.OptimizationRequest) string {
	// 清理多余的空行
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

// BuildPrompt 构建结构化优化提示词
func (s *StructureStrategy) BuildPrompt(req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildStrategyPrompt(models.StrategyStructure, req)
}

// hasGoodStructure 检查内容是否已经有良好的结构
func (s *StructureStrategy) hasGoodStructure(content string) bool {
	// 检查是否有标题
	hasHeading := strings.Contains(content, "# ") || strings.Contains(content, "## ")

	// 检查是否有列表
	hasList := strings.Contains(content, "- ") || strings.Contains(content, "* ") ||
		strings.Contains(content, "1. ") || strings.Contains(content, "2. ")

	// 检查是否有分段
	hasParagraph := strings.Contains(content, "\n\n")

	return hasHeading && (hasList || hasParagraph)
}

// GetStructureScore 获取结构化评分
func (s *StructureStrategy) GetStructureScore(content string) float64 {
	score := 0.0

	// 1. 检查标题层级 (0-30分)
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

	// 2. 检查列表使用 (0-20分)
	if strings.Contains(content, "- ") || strings.Contains(content, "* ") {
		score += 10.0
	}
	if strings.Contains(content, "1.") || strings.Contains(content, "2.") {
		score += 10.0
	}

	// 3. 检查段落结构 (0-25分)
	paragraphs := strings.Split(content, "\n\n")
	if len(paragraphs) > 1 {
		score += 15.0
	}
	if len(paragraphs) > 3 {
		score += 10.0
	}

	// 4. 检查章节划分 (0-25分)
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

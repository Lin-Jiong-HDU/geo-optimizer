package prompts

import (
	"fmt"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// Prompt 模板常量
const (
	// System prompts
	SystemPromptGEOExpert       = "你是一个专业的GEO（生成式搜索引擎优化）专家。你的任务是优化内容以提高在AI搜索引擎中的可见性和引用率。"
	SystemPromptSchemaExpert    = "你是一个Schema.org专家，精通结构化数据标记。"
	SystemPromptContentAnalyzer = "你是一个内容分析专家，擅长评估内容的GEO质量。"

	// Strategy-specific prompts
	StrategyPromptStructure = `请优化以下内容的结构，使其更符合AI搜索引擎的偏好。

优化重点：
- 添加清晰的标题层级（H1, H2, H3）
- 使用列表组织要点
- 段落不宜过长，保持在100-300字
- 添加章节划分

原始内容：
%s

请只返回优化后的内容，不要包含解释。`

	StrategyPromptSchema = `请为以下内容生成JSON-LD格式的Schema标记。

内容：
%s

Schema类型：%s

请只返回JSON格式的Schema标记，不要包含其他解释文字。`

	StrategyPromptAnswerFirst = `请将以下内容改写为"答案优先"格式。

要求：
- 在开头直接给出关键结论或答案
- 然后再展开详细解释
- 使用简练的语言

原始内容：
%s

请只返回改写后的内容。`

	StrategyPromptAuthority = `请增强以下内容的权威性。

要求：
- 添加数据支撑（可以使用示例数据，用[数据]标注）
- 引用来源（使用[来源]标注）
- 添加专业术语
- 增加案例或证据

原始内容：
%s

请只返回增强后的内容。`

	StrategyPromptFAQ = `请为以下内容生成FAQ（常见问题）部分。

要求：
- 生成3-5个常见问题
- 每个问题给出简洁准确的答案
- 问题应该覆盖内容的核心要点

原始内容：
%s

请只返回FAQ部分，格式如下：
## 常见问题

### Q1: [问题]
A: [答案]

### Q2: [问题]
A: [答案]
`
)

// BuildOptimizationPrompt 构建完整的优化提示词
func BuildOptimizationPrompt(req *models.OptimizationRequest) string {
	prompt := fmt.Sprintf(`请优化以下内容以提高其在AI搜索引擎中的可见性和引用率。

【原始内容】
标题：%s
内容：%s

【企业信息】
公司名称：%s
产品名称：%s
产品描述：%s
产品特点：%v
独特卖点：%v

【目标AI平台】%v
【关键词】%v
【优化策略】%v

请提供优化后的内容，包括：
1. 优化后的完整内容
2. 建议的Schema标记（JSON-LD格式）
3. FAQ部分
4. 内容摘要

确保内容：
- 结构清晰，有明确的标题层级
- 在开头提供关键结论
- 包含权威性引用和数据支持
- 自然地植入企业产品信息
- 针对目标AI平台的偏好进行优化`,
		req.Title,
		req.Content,
		req.Enterprise.CompanyName,
		req.Enterprise.ProductName,
		req.Enterprise.ProductDescription,
		req.Enterprise.ProductFeatures,
		req.Enterprise.USP,
		req.TargetAI,
		req.Keywords,
		req.Strategies)

	// 添加竞品信息
	if len(req.Competitors) > 0 {
		competitorInfo := "\n\n【竞品信息】\n"
		for _, comp := range req.Competitors {
			competitorInfo += fmt.Sprintf("- %s: 劣势%v\n", comp.Name, comp.Weaknesses)
		}
		prompt += competitorInfo
	}

	return prompt
}

// BuildStrategyPrompt 构建策略特定的提示词
func BuildStrategyPrompt(strategy models.StrategyType, content string, extraParams ...string) string {
	switch strategy {
	case models.StrategyStructure:
		return fmt.Sprintf(StrategyPromptStructure, content)
	case models.StrategySchema:
		schemaType := "Article"
		if len(extraParams) > 0 {
			schemaType = extraParams[0]
		}
		return fmt.Sprintf(StrategyPromptSchema, content, schemaType)
	case models.StrategyAnswerFirst:
		return fmt.Sprintf(StrategyPromptAnswerFirst, content)
	case models.StrategyAuthority:
		return fmt.Sprintf(StrategyPromptAuthority, content)
	case models.StrategyFAQ:
		return fmt.Sprintf(StrategyPromptFAQ, content)
	default:
		return content
	}
}

// GetSystemPrompt 获取系统提示词
func GetSystemPrompt(task string) string {
	switch task {
	case "schema":
		return SystemPromptSchemaExpert
	case "analysis":
		return SystemPromptContentAnalyzer
	default:
		return SystemPromptGEOExpert
	}
}

// ApplyAIPreferences 根据 AI 平台偏好调整提示词
func ApplyAIPreferences(prompt string, preferences []models.AIPreference) string {
	// 基础提示词
	enhancedPrompt := prompt

	// 应用每个偏好的设置
	for _, pref := range preferences {
		// 添加内容风格要求
		if pref.ContentStyle != "" {
			enhancedPrompt += fmt.Sprintf("\n\n内容风格要求：%s", pref.ContentStyle)
		}

		// 添加响应格式要求
		if pref.ResponseFormat != "" {
			enhancedPrompt += fmt.Sprintf("\n响应格式：%s", pref.ResponseFormat)
		}

		// 添加引用风格要求
		if pref.CitationStyle != "" {
			enhancedPrompt += fmt.Sprintf("\n引用风格：%s", pref.CitationStyle)
		}

		// 添加自定义指令
		if pref.CustomInstructions != "" {
			enhancedPrompt += fmt.Sprintf("\n特殊要求：%s", pref.CustomInstructions)
		}
	}

	return enhancedPrompt
}

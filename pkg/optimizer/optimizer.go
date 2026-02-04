package optimizer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/analyzer"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/config"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
	strategiespkg "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer/strategies"
)

// Optimizer GEO优化器
type Optimizer struct {
	llmClient     llm.LLMClient
	scorer        *analyzer.Scorer
	strategies    map[models.StrategyType]strategiespkg.Strategy
	promptBuilder *prompts.Builder
}

// New 创建优化器
func New(client llm.LLMClient) *Optimizer {
	scorer := analyzer.NewScorer(client)
	opt := &Optimizer{
		llmClient:     client,
		scorer:        scorer,
		strategies:    make(map[models.StrategyType]strategiespkg.Strategy),
		promptBuilder: prompts.NewBuilder(),
	}

	// 注册默认策略
	opt.RegisterDefaultStrategies()

	return opt
}

// RegisterStrategy 注册策略
func (o *Optimizer) RegisterStrategy(strategy strategiespkg.Strategy) {
	o.strategies[strategy.Type()] = strategy
}

// RegisterDefaultStrategies 注册默认策略
func (o *Optimizer) RegisterDefaultStrategies() {
	o.RegisterStrategy(strategiespkg.NewStructureStrategy())
	o.RegisterStrategy(strategiespkg.NewSchemaStrategy())
	o.RegisterStrategy(strategiespkg.NewAnswerFirstStrategy())
	o.RegisterStrategy(strategiespkg.NewAuthorityStrategy())
	o.RegisterStrategy(strategiespkg.NewFAQStrategy())
}

// Optimize 执行完整优化流程
func (o *Optimizer) Optimize(ctx context.Context, req *models.OptimizationRequest) (*models.OptimizationResponse, error) {
	// 1. 参数验证
	if err := o.validateRequest(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// 2. 设置默认值
	o.applyDefaults(req)

	// 3. 内容评分（优化前）
	scoreBefore, err := o.scorer.Score(ctx, req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to score content before optimization: %w", err)
	}

	// 4. 策略执行
	optimizedContent, schemaMarkup, faqSection, err := o.executeStrategies(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("strategy execution failed: %w", err)
	}

	// 5. 内容评分（优化后）
	scoreAfter, err := o.scorer.Score(ctx, optimizedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to score content after optimization: %w", err)
	}

	// 6. 解析响应
	response := o.buildResponse(req, optimizedContent, schemaMarkup, faqSection, scoreBefore, scoreAfter)

	return response, nil
}

// OptimizeWithStrategy 使用指定策略执行优化
func (o *Optimizer) OptimizeWithStrategy(ctx context.Context, req *models.OptimizationRequest, strategyType models.StrategyType) (*models.OptimizationResponse, error) {
	// 1. 参数验证
	if err := o.validateRequest(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// 2. 设置默认值
	o.applyDefaults(req)

	// 3. 获取策略
	strategy, ok := o.strategies[strategyType]
	if !ok {
		return nil, fmt.Errorf("strategy %s not registered", strategyType)
	}

	// 4. 验证策略是否适用
	if !strategy.Validate(req) {
		return nil, fmt.Errorf("strategy %s is not applicable to this request", strategyType)
	}

	// 5. 内容评分（优化前）
	scoreBefore, err := o.scorer.Score(ctx, req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to score content before optimization: %w", err)
	}

	// 6. 执行策略
	optimizedContent, err := o.executeStrategy(ctx, req, strategy)
	if err != nil {
		return nil, fmt.Errorf("strategy execution failed: %w", err)
	}

	// 7. 内容评分（优化后）
	scoreAfter, err := o.scorer.Score(ctx, optimizedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to score content after optimization: %w", err)
	}

	// 8. 构建响应
	response := &models.OptimizationResponse{
		OptimizedContent:  optimizedContent,
		Title:             req.Title,
		AppliedStrategies: []models.StrategyType{strategyType},
		ScoreBefore:       scoreBefore.OverallScore(),
		ScoreAfter:        scoreAfter.OverallScore(),
		GeneratedAt:       time.Now(),
		Version:           "1.0.0",
	}

	return response, nil
}

// executeStrategies 执行所有策略
func (o *Optimizer) executeStrategies(ctx context.Context, req *models.OptimizationRequest) (content, schema, faq string, err error) {
	content = req.Content
	var strategyResults []string

	// 遍历策略
	for _, strategyType := range req.Strategies {
		strategy, ok := o.strategies[strategyType]
		if !ok {
			continue // 跳过未注册的策略
		}

		if !strategy.Validate(req) {
			continue // 跳过不适用的策略
		}

		// 执行策略
		result, err := o.executeStrategy(ctx, req, strategy)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to execute strategy %s: %w", strategyType, err)
		}

		strategyResults = append(strategyResults, result)
	}

	// 合并策略结果
	if len(strategyResults) > 0 {
		// 使用最后一个策略的结果作为主要内容
		content = strategyResults[len(strategyResults)-1]
	}

	// 提取 Schema 和 FAQ
	schema = o.extractSchema(content)
	faq = o.extractFAQ(content)

	return content, schema, faq, nil
}

// executeStrategy 执行单个策略
func (o *Optimizer) executeStrategy(ctx context.Context, req *models.OptimizationRequest, strategy strategiespkg.Strategy) (string, error) {
	// 1. 预处理（暂不使用结果，仅调用）
	_ = strategy.Preprocess(req.Content, req)

	// 2. 构建提示词
	prompt := strategy.BuildPrompt(req)

	// 3. 调用 LLM
	messages := []llm.Message{
		{Role: "system", Content: prompts.GetSystemPrompt("optimization")},
		{Role: "user", Content: prompt},
	}

	chatResp, err := o.llmClient.Chat(ctx, &llm.ChatRequest{
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   8000,
	})
	if err != nil {
		return "", fmt.Errorf("LLM chat failed: %w", err)
	}

	// 4. 后处理
	result := strategy.Postprocess(chatResp.Content, req)

	return result, nil
}

// validateRequest 验证请求参数
func (o *Optimizer) validateRequest(req *models.OptimizationRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}

	if req.Content == "" {
		return fmt.Errorf("content is required")
	}

	if req.Title == "" {
		return fmt.Errorf("title is required")
	}

	return nil
}

// applyDefaults 应用默认值
func (o *Optimizer) applyDefaults(req *models.OptimizationRequest) {
	// 如果没有指定策略，使用默认策略
	if len(req.Strategies) == 0 {
		req.Strategies = []models.StrategyType{
			models.StrategyStructure,
			models.StrategySchema,
		}
	}

	// 如果没有指定目标AI平台，使用默认平台
	if len(req.TargetAI) == 0 {
		req.TargetAI = []string{"chatgpt"}
	}

	// 应用AI平台预设配置
	if req.AIPreferences == nil {
		req.AIPreferences = make(map[string]models.AIPreference)
	}
	for _, ai := range req.TargetAI {
		if _, exists := req.AIPreferences[ai]; !exists {
			if profile, ok := config.AIProfiles[ai]; ok {
				req.AIPreferences[ai] = profile
			}
		}
	}
}

// buildResponse 构建响应
func (o *Optimizer) buildResponse(req *models.OptimizationRequest, optimizedContent, schemaMarkup, faqSection string, scoreBefore, scoreAfter *models.GeoScore) *models.OptimizationResponse {
	// 提取摘要
	summary := o.extractSummary(optimizedContent)

	// 提取产品提及
	productMentions := o.extractProductMentions(optimizedContent, req.Enterprise.ProductName)

	// 计算提及次数
	mentionCount := o.countProductMentions(optimizedContent, req.Enterprise.ProductName)

	// 提取差异化要点
	differentiationPoints := o.extractDifferentiationPoints(optimizedContent, req.Competitors)

	return &models.OptimizationResponse{
		OptimizedContent:      optimizedContent,
		Title:                 req.Title,
		SchemaMarkup:          schemaMarkup,
		FAQSection:            faqSection,
		Summary:               summary,
		ProductMentions:       productMentions,
		MentionCount:          mentionCount,
		DifferentiationPoints: differentiationPoints,
		AppliedStrategies:     req.Strategies,
		ScoreBefore:           scoreBefore.OverallScore(),
		ScoreAfter:            scoreAfter.OverallScore(),
		GeneratedAt:           time.Now(),
		Version:               "1.0.0",
	}
}

// extractSchema 提取Schema标记
func (o *Optimizer) extractSchema(content string) string {
	// 简单实现：查找JSON-LD代码块
	if strings.Contains(content, `<script type="application/ld+json">`) {
		start := strings.Index(content, `<script type="application/ld+json">`)
		end := strings.Index(content[start:], `</script>`)
		if end > start {
			return content[start : start+end+len(`</script>`)]
		}
	}

	// 查找JSON代码块
	if strings.Contains(content, "```json") {
		start := strings.Index(content, "```json")
		end := strings.Index(content[start:], "```")
		if end > start {
			jsonContent := strings.TrimSpace(content[start+7 : start+end])
			if strings.HasPrefix(jsonContent, "{") || strings.HasPrefix(jsonContent, "[") {
				return fmt.Sprintf(`<script type="application/ld+json">%s</script>`, jsonContent)
			}
		}
	}

	return ""
}

// extractFAQ 提取FAQ部分
func (o *Optimizer) extractFAQ(content string) string {
	// 查找FAQ章节
	keywords := []string{"## 常见问题", "## FAQ", "### 常见问题", "### FAQ"}

	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			start := strings.Index(content, keyword)
			// 查找FAQ章节的结束位置
			end := len(content)
			nextHeading := strings.Index(content[start+2:], "##")
			if nextHeading > 0 && start+2+nextHeading < end {
				end = start + 2 + nextHeading
			}
			return strings.TrimSpace(content[start:end])
		}
	}

	return ""
}

// extractSummary 提取摘要
func (o *Optimizer) extractSummary(content string) string {
	// 查找摘要部分
	keywords := []string{"## 摘要", "## 总结", "## 概述", "### 摘要", "### 总结"}

	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			start := strings.Index(content, keyword)
			end := start + len(keyword)

			// 提取摘要章节的第一段
			nextNewline := strings.Index(content[end:], "\n\n")
			if nextNewline > 0 {
				return strings.TrimSpace(content[end+2 : end+2+nextNewline])
			}

			// 提取摘要章节的前200个字符
			if end+200 < len(content) {
				return strings.TrimSpace(content[end+2 : end+200])
			}
		}
	}

	// 如果没有找到摘要章节，提取前200个字符
	if len(content) > 200 {
		return strings.TrimSpace(content[:200]) + "..."
	}

	return content
}

// extractProductMentions 提取产品提及
func (o *Optimizer) extractProductMentions(content string, productName string) []models.ProductMention {
	if productName == "" {
		return []models.ProductMention{}
	}

	var mentions []models.ProductMention
	lowerContent := strings.ToLower(content)
	lowerProduct := strings.ToLower(productName)

	pos := 0
	for {
		idx := strings.Index(lowerContent[pos:], lowerProduct)
		if idx == -1 {
			break
		}

		actualPos := pos + idx

		// 获取上下文
		start := actualPos - 50
		if start < 0 {
			start = 0
		}
		end := actualPos + len(productName) + 50
		if end > len(content) {
			end = len(content)
		}

		context := content[start:end]

		// 评估影响级别
		impactLevel := "low"
		if strings.Contains(context, "推荐") || strings.Contains(context, "首选") {
			impactLevel = "high"
		} else if strings.Contains(context, "使用") || strings.Contains(context, "选择") {
			impactLevel = "medium"
		}

		mentions = append(mentions, models.ProductMention{
			Position:    actualPos,
			Context:     context,
			ImpactLevel: impactLevel,
		})

		pos = actualPos + len(productName)
	}

	return mentions
}

// countProductMentions 计算产品提及次数
func (o *Optimizer) countProductMentions(content string, productName string) int {
	if productName == "" {
		return 0
	}

	count := strings.Count(strings.ToLower(content), strings.ToLower(productName))
	return count
}

// extractDifferentiationPoints 提取差异化要点
func (o *Optimizer) extractDifferentiationPoints(content string, competitors []models.CompetitorInfo) []string {
	var points []string

	// 查找差异化关键词
	keywords := []string{"独特", "优势", "区别", "不同于", "相比"}

	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			// 简单提取包含关键词的句子
			start := strings.Index(content, keyword)
			if start >= 0 {
				sentenceStart := strings.LastIndex(content[:start], "。")
				if sentenceStart < 0 {
					sentenceStart = 0
				} else {
					sentenceStart++
				}

				sentenceEnd := strings.Index(content[start:], "。")
				if sentenceEnd < 0 {
					sentenceEnd = len(content)
				} else {
					sentenceEnd = start + sentenceEnd
				}

				sentence := strings.TrimSpace(content[sentenceStart:sentenceEnd])
				if len(sentence) > 10 {
					points = append(points, sentence)
				}
			}
		}
	}

	return points
}

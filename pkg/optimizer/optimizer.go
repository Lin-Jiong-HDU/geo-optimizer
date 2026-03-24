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

	// 2. 克隆请求并设置默认值（P0 fix: 避免修改原始请求，解决并发安全问题）
	reqCopy := o.cloneRequest(req)
	o.applyDefaults(reqCopy)

	// 3. 内容评分（优化前）
	scoreBefore, err := o.scorer.Score(ctx, reqCopy.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to score content before optimization: %w", err)
	}

	// 4. 策略执行
	optimizedContent, schemaMarkup, faqSection, totalTokens, llmModel, err := o.executeStrategies(ctx, reqCopy)
	if err != nil {
		return nil, fmt.Errorf("strategy execution failed: %w", err)
	}

	// 5. 内容评分（优化后）
	scoreAfter, err := o.scorer.Score(ctx, optimizedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to score content after optimization: %w", err)
	}

	// 6. 解析响应
	response := o.buildResponse(reqCopy, optimizedContent, schemaMarkup, faqSection, totalTokens, llmModel, scoreBefore, scoreAfter)

	return response, nil
}

// OptimizeWithStrategy 使用指定策略执行优化
func (o *Optimizer) OptimizeWithStrategy(ctx context.Context, req *models.OptimizationRequest, strategyType models.StrategyType) (*models.OptimizationResponse, error) {
	// 1. 参数验证
	if err := o.validateRequest(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// 2. 克隆请求并设置默认值（P0 fix: 避免修改原始请求，解决并发安全问题）
	reqCopy := o.cloneRequest(req)
	o.applyDefaults(reqCopy)

	// 3. 获取策略
	strategy, ok := o.strategies[strategyType]
	if !ok {
		return nil, fmt.Errorf("strategy %s not registered", strategyType)
	}

	// 4. 验证策略是否适用
	if !strategy.Validate(reqCopy) {
		return nil, fmt.Errorf("strategy %s is not applicable to this request", strategyType)
	}

	// 5. 内容评分（优化前）
	scoreBefore, err := o.scorer.Score(ctx, reqCopy.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to score content before optimization: %w", err)
	}

	// 6. 执行策略
	optimizedContent, chatResp, err := o.executeStrategy(ctx, reqCopy, strategy)
	if err != nil {
		return nil, fmt.Errorf("strategy execution failed: %w", err)
	}

	// 7. 内容评分（优化后）
	scoreAfter, err := o.scorer.Score(ctx, optimizedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to score content after optimization: %w", err)
	}

	// 8. 构建响应（使用 buildResponse 保持与 Optimize 一致）
	// 根据策略类型确定 schema 和 faq 内容
	var schemaMarkup, faqSection string
	switch strategyType {
	case models.StrategySchema:
		schemaMarkup = optimizedContent
		optimizedContent = reqCopy.Content // Schema策略不修改主体内容
	case models.StrategyFAQ:
		faqSection = optimizedContent
		optimizedContent = reqCopy.Content // FAQ策略不修改主体内容
	}

	response := o.buildResponse(reqCopy, optimizedContent, schemaMarkup, faqSection, chatResp.TokensUsed, chatResp.Model, scoreBefore, scoreAfter)
	response.AppliedStrategies = []models.StrategyType{strategyType} // 覆盖为单策略

	return response, nil
}

// executeStrategies 执行所有策略（支持累积优化）
func (o *Optimizer) executeStrategies(ctx context.Context, req *models.OptimizationRequest) (content, schema, faq string, totalTokens int, model string, err error) {
	currentContent := req.Content // 当前内容（会被更新）
	schema = ""
	faq = ""
	totalTokens = 0
	model = ""

	// 遍历策略
	for _, strategyType := range req.Strategies {
		strategy, ok := o.strategies[strategyType]
		if !ok {
			continue // 跳过未注册的策略
		}

		if !strategy.Validate(req) {
			continue // 跳过不适用的策略
		}

		// 使用新方法：传入当前内容执行策略（支持累积优化）
		result, chatResp, err := o.executeStrategyWithContent(ctx, currentContent, req, strategy)
		if err != nil {
			return "", "", "", 0, "", fmt.Errorf("failed to execute strategy %s: %w", strategyType, err)
		}

		// 累加 token 和记录模型
		totalTokens += chatResp.TokensUsed
		if model == "" && chatResp.Model != "" {
			model = chatResp.Model
		}

		// 根据策略类型处理结果
		switch strategyType {
		case models.StrategyStructure, models.StrategyAnswerFirst, models.StrategyAuthority:
			currentContent = result // 累积更新内容
		case models.StrategySchema:
			schema = result // Schema单独存储
		case models.StrategyFAQ:
			faq = result // FAQ单独存储
			// FAQ不追加到content，保持内容独立
		}
	}

	// 如果没有生成schema，尝试从最终内容中提取
	if schema == "" {
		schema = o.extractSchema(currentContent)
	}
	// 如果没有生成faq，尝试从最终内容中提取
	if faq == "" {
		faq = o.extractFAQ(currentContent)
	}

	return currentContent, schema, faq, totalTokens, model, nil
}

// executeStrategy 执行单个策略（内部方法，用于单策略优化）
func (o *Optimizer) executeStrategy(ctx context.Context, req *models.OptimizationRequest, strategy strategiespkg.Strategy) (string, *llm.ChatResponse, error) {
	_ = strategy.Preprocess(req.Content, req)

	prompt := strategy.BuildPrompt(req)

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
		return "", nil, fmt.Errorf("LLM chat failed: %w", err)
	}

	result := strategy.Postprocess(chatResp.Content, req)

	return result, chatResp, nil
}

// executeStrategyWithContent 使用指定内容执行策略（支持累积优化）
func (o *Optimizer) executeStrategyWithContent(ctx context.Context, content string, req *models.OptimizationRequest, strategy strategiespkg.Strategy) (string, *llm.ChatResponse, error) {
	// 1. 鄄处理（使用当前内容）
	preprocessed := strategy.Preprocess(content, req)

	// 2. 构建Prompt（使用预处理后的内容）
	prompt := strategy.BuildPromptWithContent(preprocessed, req)

	// 3. 调用LLM
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
		return "", nil, fmt.Errorf("LLM chat failed: %w", err)
	}

	// 4. 后处理
	result := strategy.Postprocess(chatResp.Content, req)

	return result, chatResp, nil
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

// cloneRequest 克隆请求对象，避免修改原始请求（P0 fix: 解决并发安全问题）
func (o *Optimizer) cloneRequest(req *models.OptimizationRequest) *models.OptimizationRequest {
	if req == nil {
		return nil
	}

	clone := &models.OptimizationRequest{
		Content:               req.Content,
		Title:                 req.Title,
		URL:                   req.URL,
		Enterprise:            req.Enterprise,
		TargetAI:              make([]string, len(req.TargetAI)),
		Keywords:              make([]string, len(req.Keywords)),
		Strategies:            make([]models.StrategyType, len(req.Strategies)),
		Tone:                  req.Tone,
		Industry:              req.Industry,
		IncludeProductMention: req.IncludeProductMention,
		MentionFrequency:      req.MentionFrequency,
		CallToAction:          req.CallToAction,
		Author:                req.Author,
		PublishDate:           req.PublishDate,
	}

	// 复制切片
	copy(clone.TargetAI, req.TargetAI)
	copy(clone.Keywords, req.Keywords)
	copy(clone.Strategies, req.Strategies)

	// 复制竞品信息
	if req.Competitors != nil {
		clone.Competitors = make([]models.CompetitorInfo, len(req.Competitors))
		copy(clone.Competitors, req.Competitors)
	}

	// 复制 AI 偏好配置（创建新 map 避免并发写入问题）
	clone.AIPreferences = make(map[string]models.AIPreference)
	for k, v := range req.AIPreferences {
		clone.AIPreferences[k] = v
	}

	// 复制自定义数据
	if req.CustomData != nil {
		clone.CustomData = make(map[string]string)
		for k, v := range req.CustomData {
			clone.CustomData[k] = v
		}
	}

	return clone
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
func (o *Optimizer) buildResponse(req *models.OptimizationRequest, optimizedContent, schemaMarkup, faqSection string, totalTokens int, llmModel string, scoreBefore, scoreAfter *models.GeoScore) *models.OptimizationResponse {
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
		Recommendations:       o.generateRecommendations(scoreBefore, scoreAfter),
		GeneratedAt:           time.Now(),
		Version:               "1.0.0",
		LLMModel:              llmModel,
		TokensUsed:            totalTokens,
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

// generateRecommendations 生成改进建议
func (o *Optimizer) generateRecommendations(before, after *models.GeoScore) []string {
	var recommendations []string

	// 基于优化后评分生成建议
	if after.Structure < 60 {
		recommendations = append(recommendations, "建议添加更多标题层级和列表来改善内容结构")
	}
	if after.Authority < 60 {
		recommendations = append(recommendations, "建议添加数据支撑和引用来源来提升权威性")
	}
	if after.Clarity < 60 {
		recommendations = append(recommendations, "建议简化句子结构，使用更通俗的表达来提升清晰度")
	}
	if after.Citation < 60 {
		recommendations = append(recommendations, "建议增加关键信息和引用点来提升可引用性")
	}
	if after.Schema < 60 {
		recommendations = append(recommendations, "建议完善结构化数据标记以提升AI可理解性")
	}

	// 如果没有建议，返回积极反馈
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "内容优化效果良好，各项指标均达到较高水平")
	}

	return recommendations
}

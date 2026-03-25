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

// Optimizer performs GEO optimization on content.
type Optimizer struct {
	llmClient     llm.LLMClient
	scorer        *analyzer.Scorer
	strategies    map[models.StrategyType]strategiespkg.Strategy
	promptBuilder *prompts.Builder
}

// New creates a new Optimizer instance.
func New(client llm.LLMClient) *Optimizer {
	scorer := analyzer.NewScorer(client)
	opt := &Optimizer{
		llmClient:     client,
		scorer:        scorer,
		strategies:    make(map[models.StrategyType]strategiespkg.Strategy),
		promptBuilder: prompts.NewBuilder(),
	}

	opt.RegisterDefaultStrategies()

	return opt
}

// RegisterStrategy registers a strategy with the optimizer.
func (o *Optimizer) RegisterStrategy(strategy strategiespkg.Strategy) {
	o.strategies[strategy.Type()] = strategy
}

// RegisterDefaultStrategies registers the default optimization strategies.
func (o *Optimizer) RegisterDefaultStrategies() {
	o.RegisterStrategy(strategiespkg.NewStructureStrategy())
	o.RegisterStrategy(strategiespkg.NewSchemaStrategy())
	o.RegisterStrategy(strategiespkg.NewAnswerFirstStrategy())
	o.RegisterStrategy(strategiespkg.NewAuthorityStrategy())
	o.RegisterStrategy(strategiespkg.NewFAQStrategy())
}

// Optimize executes the full optimization pipeline.
func (o *Optimizer) Optimize(ctx context.Context, req *models.OptimizationRequest) (*models.OptimizationResponse, error) {
	if err := o.validateRequest(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	reqCopy := o.cloneRequest(req)
	o.applyDefaults(reqCopy)

	scoreBefore, err := o.scorer.Score(ctx, reqCopy.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to score content before optimization: %w", err)
	}

	optimizedContent, schemaMarkup, faqSection, totalTokens, llmModel, err := o.executeStrategies(ctx, reqCopy)
	if err != nil {
		return nil, fmt.Errorf("strategy execution failed: %w", err)
	}

	scoreAfter, err := o.scorer.Score(ctx, optimizedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to score content after optimization: %w", err)
	}

	response := o.buildResponse(reqCopy, optimizedContent, schemaMarkup, faqSection, totalTokens, llmModel, scoreBefore, scoreAfter)

	return response, nil
}

// OptimizeWithStrategy executes optimization with a specific strategy.
func (o *Optimizer) OptimizeWithStrategy(ctx context.Context, req *models.OptimizationRequest, strategyType models.StrategyType) (*models.OptimizationResponse, error) {
	if err := o.validateRequest(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	reqCopy := o.cloneRequest(req)
	o.applyDefaults(reqCopy)

	strategy, ok := o.strategies[strategyType]
	if !ok {
		return nil, fmt.Errorf("strategy %s not registered", strategyType)
	}

	if !strategy.Validate(reqCopy) {
		return nil, fmt.Errorf("strategy %s is not applicable to this request", strategyType)
	}

	scoreBefore, err := o.scorer.Score(ctx, reqCopy.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to score content before optimization: %w", err)
	}

	strategyResult, chatResp, err := o.executeStrategy(ctx, reqCopy, strategy)
	if err != nil {
		return nil, fmt.Errorf("strategy execution failed: %w", err)
	}

	var optimizedContent, schemaMarkup, faqSection string
	switch strategyType {
	case models.StrategySchema:
		schemaMarkup = strategyResult
		optimizedContent = reqCopy.Content
	case models.StrategyFAQ:
		faqSection = strategyResult
		optimizedContent = reqCopy.Content
	default:
		optimizedContent = strategyResult
	}

	scoreAfter, err := o.scorer.Score(ctx, optimizedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to score content after optimization: %w", err)
	}

	response := o.buildResponse(reqCopy, optimizedContent, schemaMarkup, faqSection, chatResp.TokensUsed, chatResp.Model, scoreBefore, scoreAfter)
	response.AppliedStrategies = []models.StrategyType{strategyType}

	return response, nil
}

// executeStrategies executes all strategies with cumulative optimization support.
func (o *Optimizer) executeStrategies(ctx context.Context, req *models.OptimizationRequest) (content, schema, faq string, totalTokens int, model string, err error) {
	currentContent := req.Content
	schema = ""
	faq = ""
	totalTokens = 0
	model = ""

	for _, strategyType := range req.Strategies {
		strategy, ok := o.strategies[strategyType]
		if !ok {
			continue
		}

		if !strategy.Validate(req) {
			continue
		}

		result, chatResp, err := o.executeStrategyWithContent(ctx, currentContent, req, strategy)
		if err != nil {
			return "", "", "", 0, "", fmt.Errorf("failed to execute strategy %s: %w", strategyType, err)
		}

		totalTokens += chatResp.TokensUsed
		if model == "" && chatResp.Model != "" {
			model = chatResp.Model
		}

		switch strategyType {
		case models.StrategyStructure, models.StrategyAnswerFirst, models.StrategyAuthority:
			currentContent = result
		case models.StrategySchema:
			schema = result
		case models.StrategyFAQ:
			faq = result
		}
	}

	if schema == "" {
		schema = o.extractSchema(currentContent)
	}
	if faq == "" {
		faq = o.extractFAQ(currentContent)
	}

	return currentContent, schema, faq, totalTokens, model, nil
}

// executeStrategy executes a single strategy.
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

// executeStrategyWithContent executes a strategy with specified content for cumulative optimization.
func (o *Optimizer) executeStrategyWithContent(ctx context.Context, content string, req *models.OptimizationRequest, strategy strategiespkg.Strategy) (string, *llm.ChatResponse, error) {
	preprocessed := strategy.Preprocess(content, req)

	prompt := strategy.BuildPromptWithContent(preprocessed, req)

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

// validateRequest validates the optimization request.
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

// cloneRequest creates a deep copy of the request to avoid modifying the original.
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

	copy(clone.TargetAI, req.TargetAI)
	copy(clone.Keywords, req.Keywords)
	copy(clone.Strategies, req.Strategies)

	if req.Competitors != nil {
		clone.Competitors = make([]models.CompetitorInfo, len(req.Competitors))
		copy(clone.Competitors, req.Competitors)
	}

	clone.AIPreferences = make(map[string]models.AIPreference)
	for k, v := range req.AIPreferences {
		clone.AIPreferences[k] = v
	}

	if req.CustomData != nil {
		clone.CustomData = make(map[string]string)
		for k, v := range req.CustomData {
			clone.CustomData[k] = v
		}
	}

	return clone
}

// applyDefaults applies default values to the request.
func (o *Optimizer) applyDefaults(req *models.OptimizationRequest) {
	if len(req.Strategies) == 0 {
		req.Strategies = []models.StrategyType{
			models.StrategyStructure,
			models.StrategySchema,
		}
	}

	if len(req.TargetAI) == 0 {
		req.TargetAI = []string{"chatgpt"}
	}

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

// buildResponse constructs the optimization response.
func (o *Optimizer) buildResponse(req *models.OptimizationRequest, optimizedContent, schemaMarkup, faqSection string, totalTokens int, llmModel string, scoreBefore, scoreAfter *models.GeoScore) *models.OptimizationResponse {
	summary := o.extractSummary(optimizedContent)

	productMentions := o.extractProductMentions(optimizedContent, req.Enterprise.ProductName)

	mentionCount := o.countProductMentions(optimizedContent, req.Enterprise.ProductName)

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

// extractSchema extracts Schema markup from content.
func (o *Optimizer) extractSchema(content string) string {
	if strings.Contains(content, `<script type="application/ld+json">`) {
		start := strings.Index(content, `<script type="application/ld+json">`)
		end := strings.Index(content[start:], `</script>`)
		if end > start {
			return content[start : start+end+len(`</script>`)]
		}
	}

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

// extractFAQ extracts the FAQ section from content.
func (o *Optimizer) extractFAQ(content string) string {
	keywords := []string{"## FAQ", "## Frequently Asked Questions", "## 常见问题", "### FAQ", "### 常见问题"}

	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			start := strings.Index(content, keyword)
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

// extractSummary extracts the summary section from content.
func (o *Optimizer) extractSummary(content string) string {
	keywords := []string{"## Summary", "## Conclusion", "## Overview", "## 摘要", "## 总结", "## 概述"}

	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			start := strings.Index(content, keyword)
			end := start + len(keyword)

			nextNewline := strings.Index(content[end:], "\n\n")
			if nextNewline > 0 {
				return strings.TrimSpace(content[end+2 : end+2+nextNewline])
			}

			if end+200 < len(content) {
				return strings.TrimSpace(content[end+2 : end+200])
			}
		}
	}

	if len(content) > 200 {
		return strings.TrimSpace(content[:200]) + "..."
	}

	return content
}

// extractProductMentions extracts product mentions from content.
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

		start := actualPos - 50
		if start < 0 {
			start = 0
		}
		end := actualPos + len(productName) + 50
		if end > len(content) {
			end = len(content)
		}

		context := content[start:end]

		impactLevel := "low"
		if strings.Contains(context, "recommend") || strings.Contains(context, "首选") || strings.Contains(context, "recommended") {
			impactLevel = "high"
		} else if strings.Contains(context, "use") || strings.Contains(context, "使用") || strings.Contains(context, "选择") {
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

// countProductMentions counts the number of product mentions in content.
func (o *Optimizer) countProductMentions(content string, productName string) int {
	if productName == "" {
		return 0
	}

	count := strings.Count(strings.ToLower(content), strings.ToLower(productName))
	return count
}

// extractDifferentiationPoints extracts differentiation points from content.
func (o *Optimizer) extractDifferentiationPoints(content string, competitors []models.CompetitorInfo) []string {
	var points []string

	keywords := []string{"unique", "advantage", "different", "compared to", "独特", "优势", "区别", "不同于", "相比"}

	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
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

// generateRecommendations generates improvement recommendations based on scores.
func (o *Optimizer) generateRecommendations(before, after *models.GeoScore) []string {
	var recommendations []string

	if after.Structure < 60 {
		recommendations = append(recommendations, "Consider adding more heading levels and lists to improve content structure")
	}
	if after.Authority < 60 {
		recommendations = append(recommendations, "Consider adding data support and citation sources to enhance authority")
	}
	if after.Clarity < 60 {
		recommendations = append(recommendations, "Consider simplifying sentence structure and using clearer expressions")
	}
	if after.Citation < 60 {
		recommendations = append(recommendations, "Consider adding key information and citation points to improve citability")
	}
	if after.Schema < 60 {
		recommendations = append(recommendations, "Consider improving structured data markup to enhance AI understanding")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Content optimization is effective, all metrics have reached a good level")
	}

	return recommendations
}

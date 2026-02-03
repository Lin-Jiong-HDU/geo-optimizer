package analyzer

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// Scorer 内容评分器
type Scorer struct {
	llmClient llm.LLMClient
}

// NewScorer 创建评分器
func NewScorer(client llm.LLMClient) *Scorer {
	return &Scorer{
		llmClient: client,
	}
}

// Score 对内容进行GEO评分（基于规则的快速评分）
func (s *Scorer) Score(ctx context.Context, content string) (*models.GeoScore, error) {
	return s.scoreByRules(content), nil
}

// ScoreWithAnalysis 对内容进行评分并返回详细分析（基于LLM）
func (s *Scorer) ScoreWithAnalysis(ctx context.Context, content string) (*models.ContentAnalysis, error) {
	llmAnalysis, err := s.llmClient.AnalyzeContent(ctx, content)
	if err != nil {
		return nil, err
	}

	// 转换为 models.ContentAnalysis
	analysis := &models.ContentAnalysis{
		StructureScore: llmAnalysis.StructureScore,
		AuthorityScore: llmAnalysis.AuthorityScore,
		ClarityScore:   llmAnalysis.ClarityScore,
		CitationScore:  llmAnalysis.CitationScore,
		SchemaScore:    llmAnalysis.SchemaScore,
		GeoScore:       llmAnalysis.TotalScore,
		Suggestions:    llmAnalysis.Suggestions,
	}

	return analysis, nil
}

// Compare 对比优化前后的评分
func (s *Scorer) Compare(ctx context.Context, before, after string) (*ScoreComparison, error) {
	// 评分前
	scoreBefore, err := s.Score(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("failed to score before content: %w", err)
	}

	// 评分后
	scoreAfter, err := s.Score(ctx, after)
	if err != nil {
		return nil, fmt.Errorf("failed to score after content: %w", err)
	}

	// 计算提升幅度
	improvements := map[string]float64{
		"structure": scoreAfter.Structure - scoreBefore.Structure,
		"authority": scoreAfter.Authority - scoreBefore.Authority,
		"clarity":   scoreAfter.Clarity - scoreBefore.Clarity,
		"citation":  scoreAfter.Citation - scoreBefore.Citation,
		"schema":    scoreAfter.Schema - scoreBefore.Schema,
	}

	totalChange := scoreAfter.OverallScore() - scoreBefore.OverallScore()

	return &ScoreComparison{
		Before:       scoreBefore,
		After:        scoreAfter,
		Improvements: improvements,
		TotalChange:  totalChange,
	}, nil
}

// scoreByRules 基于规则的内容评分
func (s *Scorer) scoreByRules(content string) *models.GeoScore {
	score := &models.GeoScore{
		Structure: s.calculateStructureScore(content),
		Authority: s.calculateAuthorityScore(content),
		Clarity:   s.calculateClarityScore(content),
		Citation:  s.calculateCitationScore(content),
		Schema:    s.calculateSchemaScore(content),
	}

	return score
}

// calculateStructureScore 计算结构化评分
func (s *Scorer) calculateStructureScore(content string) float64 {
	score := 0.0

	// 1. 检查标题层级 (0-30分)
	headingPattern := regexp.MustCompile(`(?m)^#{1,6}\s+.+`)
	headings := headingPattern.FindAllString(content, -1)
	if len(headings) > 0 {
		score += 15.0
		if len(headings) >= 3 {
			score += 10.0
		}
		if len(headings) >= 5 {
			score += 5.0
		}
	}

	// 2. 检查列表使用 (0-20分)
	if strings.Contains(content, "- ") || strings.Contains(content, "* ") {
		score += 10.0
	}
	if strings.Contains(content, "1.") || regexp.MustCompile(`^\d+\.\s`).MatchString(content) {
		score += 10.0
	}

	// 3. 检查段落结构 (0-25分)
	paragraphs := strings.Split(content, "\n\n")
	wellStructuredParagraphs := 0
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p != "" && utf8.RuneCountInString(p) <= 500 {
			wellStructuredParagraphs++
		}
	}
	if len(paragraphs) > 0 {
		ratio := float64(wellStructuredParagraphs) / float64(len(paragraphs))
		score += ratio * 25.0
	}

	// 4. 检查章节划分 (0-25分)
	sections := regexp.MustCompile(`(?m)^#{2,}\s+.+`).FindAllString(content, -1)
	if len(sections) >= 2 {
		score += 15.0
	}
	if len(sections) >= 4 {
		score += 10.0
	}

	// 确保分数在 0-100 范围内
	if score > 100 {
		score = 100
	}

	return score
}

// calculateAuthorityScore 计算权威性评分
func (s *Scorer) calculateAuthorityScore(content string) float64 {
	score := 0.0

	// 1. 检查数据支撑 (0-30分)
	dataPatterns := []string{
		`\d+%\s`,         // 百分比
		`\d+\s*(万|千|百)`,  // 中文数字
		`\d{4}\s*年`,      // 年份
		`according\s+to`, // "根据"
		`数据显示`,           // 中文
		`研究`,             // 研究
		`报告`,             // 报告
	}
	for _, pattern := range dataPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			score += 5.0
		}
	}
	if score > 30 {
		score = 30
	}

	// 2. 检查来源标注 (0-30分)
	sourcePatterns := []string{
		`来源[：:]\s*\w+`,
		`据\s*\w+报道`,
		`\[\d+\]`,        // 引用标记
		`\([^)]+\d{4}\)`, // 学术引用 (Author, 2024)
		`来源:\s*\[`,       // Markdown链接
	}
	sourceCount := 0
	for _, pattern := range sourcePatterns {
		matches := regexp.MustCompile(pattern).FindAllString(content, -1)
		sourceCount += len(matches)
	}
	score += float64(sourceCount) * 5.0
	if score > 30 {
		score = 30
	}

	// 3. 检查专业术语使用 (0-20分)
	professionalWords := []string{
		`优化`, `策略`, `分析`, `评估`, `效果`,
		`optimize`, `strategy`, `analysis`, `evaluation`,
	}
	termCount := 0
	for _, word := range professionalWords {
		if strings.Contains(content, word) {
			termCount++
		}
	}
	score += float64(termCount) * 4.0
	if score > 20 {
		score = 20
	}

	// 4. 检查具体案例或证据 (0-20分)
	evidencePatterns := []string{
		`例如`, `比如`, `案例`, `举例`,
		`for\s+example`, `for\s+instance`, `case\s+study`,
	}
	evidenceCount := 0
	for _, pattern := range evidencePatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			evidenceCount++
		}
	}
	score += float64(evidenceCount) * 5.0
	if score > 20 {
		score = 20
	}

	if score > 100 {
		score = 100
	}

	return score
}

// calculateClarityScore 计算清晰度评分
func (s *Scorer) calculateClarityScore(content string) float64 {
	// 空内容直接返回0
	if strings.TrimSpace(content) == "" {
		return 0.0
	}

	score := 0.0

	// 1. 检查句子长度 (0-30分)
	sentences := regexp.MustCompile(`[。！？.!?]`).Split(content, -1)
	shortSentences := 0
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s != "" && utf8.RuneCountInString(s) <= 100 {
			shortSentences++
		}
	}
	if len(sentences) > 0 {
		ratio := float64(shortSentences) / float64(len(sentences))
		score += ratio * 30.0
	}

	// 2. 检查段落长度 (0-25分)
	paragraphs := strings.Split(content, "\n\n")
	shortParagraphs := 0
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p != "" && utf8.RuneCountInString(p) <= 300 {
			shortParagraphs++
		}
	}
	if len(paragraphs) > 0 {
		ratio := float64(shortParagraphs) / float64(len(paragraphs))
		score += ratio * 25.0
	}

	// 3. 检查逻辑连接词 (0-25分)
	connectorPatterns := []string{
		`因此`, `所以`, `但是`, `然而`, `此外`, `另外`,
		`therefore`, `however`, `moreover`, `furthermore`, `in addition`,
	}
	connectorCount := 0
	for _, pattern := range connectorPatterns {
		if strings.Contains(content, pattern) {
			connectorCount++
		}
	}
	score += float64(connectorCount) * 5.0
	if score > 25 {
		score = 25
	}

	// 4. 检查可读性（避免过多嵌套）(0-20分)
	lines := strings.Split(content, "\n")
	longLines := 0
	for _, line := range lines {
		if utf8.RuneCountInString(line) > 100 {
			longLines++
		}
	}
	if len(lines) > 0 {
		longLineRatio := float64(longLines) / float64(len(lines))
		score += (1 - longLineRatio) * 20.0
	}

	if score > 100 {
		score = 100
	}

	return score
}

// calculateCitationScore 计算可引用性评分
func (s *Scorer) calculateCitationScore(content string) float64 {
	score := 0.0

	// 1. 检查结论性陈述 (0-30分)
	conclusionPatterns := []string{
		`结论[是为][:：].+`,
		`总结[是为][:：].+`,
		`总之`,
		`因此`,
		`in\s+conclusion`, `to\s+sum\s+up`, `therefore`,
	}
	conclusionCount := 0
	for _, pattern := range conclusionPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			conclusionCount++
		}
	}
	score += float64(conclusionCount) * 10.0
	if score > 30 {
		score = 30
	}

	// 2. 检查事实陈述 (0-30分)
	factPatterns := []string{
		`\d{4}\s*年[是以是].+`,
		`\S+\s*[是为]\s*\d+%\s*[的以上以下]`,
		`研究[显示表明].+`,
	}
	factCount := 0
	for _, pattern := range factPatterns {
		matches := regexp.MustCompile(pattern).FindAllString(content, -1)
		factCount += len(matches)
	}
	score += float64(factCount) * 5.0
	if score > 30 {
		score = 30
	}

	// 3. 检查可操作性建议 (0-20分)
	actionablePatterns := []string{
		`建议[:：]?`,
		`可以[:：]?`,
		`推荐[:：]?`,
		`should\s+\S+`, `recommend`, `suggest`,
	}
	actionableCount := 0
	for _, pattern := range actionablePatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			actionableCount++
		}
	}
	score += float64(actionableCount) * 5.0
	if score > 20 {
		score = 20
	}

	// 4. 检查独特观点 (0-20分)
	uniquePatterns := []string{
		`独特`, `创新`, `新颖`, `首次`,
		`unique`, `innovative`, `novel`, `first`,
	}
	uniqueCount := 0
	for _, pattern := range uniquePatterns {
		if strings.Contains(content, pattern) {
			uniqueCount++
		}
	}
	score += float64(uniqueCount) * 5.0
	if score > 20 {
		score = 20
	}

	if score > 100 {
		score = 100
	}

	return score
}

// calculateSchemaScore 计算Schema完整性评分
func (s *Scorer) calculateSchemaScore(content string) float64 {
	score := 0.0

	// 1. 检查JSON-LD标记 (0-50分)
	if strings.Contains(content, `<script type="application/ld+json">`) {
		score += 50.0
	} else if strings.Contains(content, `"@context"`) || strings.Contains(content, `"@type"`) {
		// 纯JSON格式
		score += 40.0
	}

	// 2. 检查Schema.org类型 (0-20分)
	schemaTypes := []string{
		`"Article"`, `"WebPage"`, `"Product"`, `"HowTo"`,
		`"FAQPage"`, `"Organization"`, `"Person"`,
	}
	for _, schemaType := range schemaTypes {
		if strings.Contains(content, schemaType) {
			score += 20.0
			break
		}
	}
	if score > 70 {
		score = 70
	}

	// 3. 检查必需字段 (0-30分)
	requiredFields := []string{
		`"name"`, `"headline"`, `"description"`,
	}
	fieldCount := 0
	for _, field := range requiredFields {
		if strings.Contains(content, field) {
			fieldCount++
		}
	}
	score += float64(fieldCount) * 10.0

	if score > 100 {
		score = 100
	}

	return score
}

// ScoreComparison 评分对比
type ScoreComparison struct {
	Before       *models.GeoScore   `json:"before"`
	After        *models.GeoScore   `json:"after"`
	Improvements map[string]float64 `json:"improvements"` // 各维度提升幅度
	TotalChange  float64            `json:"total_change"`
}

// GetImprovementSummary 获取改进摘要
func (c *ScoreComparison) GetImprovementSummary() string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("总分变化: %.2f → %.2f (%.2f%%)\n",
		c.Before.OverallScore(),
		c.After.OverallScore(),
		c.TotalChange))

	for dim, improvement := range c.Improvements {
		if improvement > 0 {
			summary.WriteString(fmt.Sprintf("%s: +%.2f\n", dim, improvement))
		} else if improvement < 0 {
			summary.WriteString(fmt.Sprintf("%s: %.2f\n", dim, improvement))
		}
	}

	return summary.String()
}

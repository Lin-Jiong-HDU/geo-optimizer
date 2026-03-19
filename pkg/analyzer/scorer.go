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

// Pre-compiled regex patterns for performance (P0 fix: avoid repeated compilation)
var (
	// Structure scoring patterns
	headingPattern     = regexp.MustCompile(`(?m)^#{1,6}\s+.+`)
	sectionPattern     = regexp.MustCompile(`(?m)^#{2,}\s+.+`)
	orderedListPattern = regexp.MustCompile(`(?m)^\d+\.\s`)

	// Authority scoring patterns
	dataPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\d+%\s`),
		regexp.MustCompile(`\d+\s*(万|千|百)`),
		regexp.MustCompile(`\d{4}\s*年`),
		regexp.MustCompile(`(?i)according\s+to`),
		regexp.MustCompile(`数据显示`),
		regexp.MustCompile(`研究`),
		regexp.MustCompile(`报告`),
	}
	sourcePatterns = []*regexp.Regexp{
		regexp.MustCompile(`来源[：:]\s*\w+`),
		regexp.MustCompile(`据\s*\w+报道`),
		regexp.MustCompile(`\[\d+\]`),
		regexp.MustCompile(`\([^)]+\d{4}\)`),
		regexp.MustCompile(`来源:\s*\[`),
	}
	evidencePatterns = []*regexp.Regexp{
		regexp.MustCompile(`例如`),
		regexp.MustCompile(`比如`),
		regexp.MustCompile(`案例`),
		regexp.MustCompile(`举例`),
		regexp.MustCompile(`(?i)for\s+example`),
		regexp.MustCompile(`(?i)for\s+instance`),
		regexp.MustCompile(`(?i)case\s+study`),
	}

	// Clarity scoring patterns
	sentenceSplitPattern = regexp.MustCompile(`[。！？.!?]`)

	// Citation scoring patterns
	conclusionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`结论[是为][:：].+`),
		regexp.MustCompile(`总结[是为][:：].+`),
		regexp.MustCompile(`总之`),
		regexp.MustCompile(`因此`),
		regexp.MustCompile(`(?i)in\s+conclusion`),
		regexp.MustCompile(`(?i)to\s+sum\s+up`),
		regexp.MustCompile(`(?i)therefore`),
	}
	factPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\d{4}\s*年[是以是].+`),
		regexp.MustCompile(`\S+\s*[是为]\s*\d+%\s*[的以上以下]`),
		regexp.MustCompile(`研究[显示表明].+`),
	}
	actionablePatterns = []*regexp.Regexp{
		regexp.MustCompile(`建议[:：]?`),
		regexp.MustCompile(`可以[:：]?`),
		regexp.MustCompile(`推荐[:：]?`),
		regexp.MustCompile(`(?i)should\s+\S+`),
		regexp.MustCompile(`(?i)recommend`),
		regexp.MustCompile(`(?i)suggest`),
	}
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
// Deprecated: The AnalyzeContent method has been moved. Use Score() instead for rule-based scoring.
func (s *Scorer) ScoreWithAnalysis(ctx context.Context, content string) (*models.ContentAnalysis, error) {
	// Use rule-based scoring instead
	score := s.scoreByRules(content)

	// Convert to ContentAnalysis
	analysis := &models.ContentAnalysis{
		StructureScore: score.Structure,
		AuthorityScore: score.Authority,
		ClarityScore:   score.Clarity,
		CitationScore:  score.Citation,
		SchemaScore:    score.Schema,
		GeoScore:       score.OverallScore(),
		Suggestions:    []string{}, // Can be enhanced later
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

	// 1. 检查标题层级 (0-30分) - 使用预编译正则
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
	if strings.Contains(content, "1.") || orderedListPattern.MatchString(content) {
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

	// 4. 检查章节划分 (0-25分) - 使用预编译正则
	sections := sectionPattern.FindAllString(content, -1)
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

	// 1. 检查数据支撑 (0-30分) - 使用预编译正则
	for _, re := range dataPatterns {
		if re.MatchString(content) {
			score += 5.0
		}
	}
	if score > 30 {
		score = 30
	}

	// 2. 检查来源标注 (0-30分) - 使用预编译正则
	sourceCount := 0
	for _, re := range sourcePatterns {
		matches := re.FindAllString(content, -1)
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

	// 4. 检查具体案例或证据 (0-20分) - 使用预编译正则
	evidenceCount := 0
	for _, re := range evidencePatterns {
		if re.MatchString(content) {
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

	// 1. 检查句子长度 (0-30分) - 使用预编译正则
	sentences := sentenceSplitPattern.Split(content, -1)
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

	// 1. 检查结论性陈述 (0-30分) - 使用预编译正则
	conclusionCount := 0
	for _, re := range conclusionPatterns {
		if re.MatchString(content) {
			conclusionCount++
		}
	}
	score += float64(conclusionCount) * 10.0
	if score > 30 {
		score = 30
	}

	// 2. 检查事实陈述 (0-30分) - 使用预编译正则
	factCount := 0
	for _, re := range factPatterns {
		matches := re.FindAllString(content, -1)
		factCount += len(matches)
	}
	score += float64(factCount) * 5.0
	if score > 30 {
		score = 30
	}

	// 3. 检查可操作性建议 (0-20分) - 使用预编译正则
	actionableCount := 0
	for _, re := range actionablePatterns {
		if re.MatchString(content) {
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

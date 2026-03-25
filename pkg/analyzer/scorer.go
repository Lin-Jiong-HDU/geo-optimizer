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

// Pre-compiled regex patterns for performance.
var (
	headingPattern     = regexp.MustCompile(`(?m)^#{1,6}\s+.+`)
	sectionPattern     = regexp.MustCompile(`(?m)^#{2,}\s+.+`)
	orderedListPattern = regexp.MustCompile(`(?m)^\d+\.\s`)

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

	sentenceSplitPattern = regexp.MustCompile(`[。！？.!?]`)

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

// Scorer performs GEO scoring on content.
type Scorer struct {
	llmClient llm.LLMClient
}

// NewScorer creates a new Scorer instance.
func NewScorer(client llm.LLMClient) *Scorer {
	return &Scorer{
		llmClient: client,
	}
}

// Score performs rule-based GEO scoring on content.
func (s *Scorer) Score(ctx context.Context, content string) (*models.GeoScore, error) {
	return s.scoreByRules(content), nil
}

// ScoreWithAnalysis performs scoring with detailed analysis.
// Deprecated: Use Score() instead for rule-based scoring.
func (s *Scorer) ScoreWithAnalysis(ctx context.Context, content string) (*models.ContentAnalysis, error) {
	score := s.scoreByRules(content)

	analysis := &models.ContentAnalysis{
		StructureScore: score.Structure,
		AuthorityScore: score.Authority,
		ClarityScore:   score.Clarity,
		CitationScore:  score.Citation,
		SchemaScore:    score.Schema,
		GeoScore:       score.OverallScore(),
		Suggestions:    []string{},
	}

	return analysis, nil
}

// Compare compares scores before and after optimization.
func (s *Scorer) Compare(ctx context.Context, before, after string) (*ScoreComparison, error) {
	scoreBefore, err := s.Score(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("failed to score before content: %w", err)
	}

	scoreAfter, err := s.Score(ctx, after)
	if err != nil {
		return nil, fmt.Errorf("failed to score after content: %w", err)
	}

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

// scoreByRules performs rule-based content scoring.
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

// calculateStructureScore calculates the structure score (0-100).
func (s *Scorer) calculateStructureScore(content string) float64 {
	score := 0.0

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

	if strings.Contains(content, "- ") || strings.Contains(content, "* ") {
		score += 10.0
	}
	if strings.Contains(content, "1.") || orderedListPattern.MatchString(content) {
		score += 10.0
	}

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

	sections := sectionPattern.FindAllString(content, -1)
	if len(sections) >= 2 {
		score += 15.0
	}
	if len(sections) >= 4 {
		score += 10.0
	}

	if score > 100 {
		score = 100
	}

	return score
}

// calculateAuthorityScore calculates the authority score (0-100).
func (s *Scorer) calculateAuthorityScore(content string) float64 {
	score := 0.0

	for _, re := range dataPatterns {
		if re.MatchString(content) {
			score += 5.0
		}
	}
	if score > 30 {
		score = 30
	}

	sourceCount := 0
	for _, re := range sourcePatterns {
		matches := re.FindAllString(content, -1)
		sourceCount += len(matches)
	}
	score += float64(sourceCount) * 5.0
	if score > 30 {
		score = 30
	}

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

// calculateClarityScore calculates the clarity score (0-100).
func (s *Scorer) calculateClarityScore(content string) float64 {
	if strings.TrimSpace(content) == "" {
		return 0.0
	}

	score := 0.0

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

// calculateCitationScore calculates the citation score (0-100).
func (s *Scorer) calculateCitationScore(content string) float64 {
	score := 0.0

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

	factCount := 0
	for _, re := range factPatterns {
		matches := re.FindAllString(content, -1)
		factCount += len(matches)
	}
	score += float64(factCount) * 5.0
	if score > 30 {
		score = 30
	}

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

// calculateSchemaScore calculates the Schema completeness score (0-100).
func (s *Scorer) calculateSchemaScore(content string) float64 {
	score := 0.0

	if strings.Contains(content, `<script type="application/ld+json">`) {
		score += 50.0
	} else if strings.Contains(content, `"@context"`) || strings.Contains(content, `"@type"`) {
		score += 40.0
	}

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

// ScoreComparison represents a comparison of scores before and after optimization.
type ScoreComparison struct {
	Before       *models.GeoScore   `json:"before"`
	After        *models.GeoScore   `json:"after"`
	Improvements map[string]float64 `json:"improvements"`
	TotalChange  float64            `json:"total_change"`
}

// GetImprovementSummary returns a summary of improvements.
func (c *ScoreComparison) GetImprovementSummary() string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("Total Score Change: %.2f → %.2f (%.2f%%)\n",
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

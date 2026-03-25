package llm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Parser handles parsing of LLM responses.
type Parser struct{}

// NewParser creates a new Parser instance.
func NewParser() *Parser {
	return &Parser{}
}

// ParseOptimizationResponse parses an optimization response.
func (p *Parser) ParseOptimizationResponse(content string) (*ParseResult, error) {
	result := &ParseResult{
		OptimizedContent: content,
		Sections:         make(map[string]string),
	}

	if jsonPart := p.extractJSON(content); jsonPart != "" {
		result.JSONData = jsonPart

		var structuredData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonPart), &structuredData); err == nil {
			result.StructuredData = structuredData
		}
	}

	result.Sections["summary"] = p.extractSection(content, []string{"Summary", "Overview", "摘要", "概述"})
	result.Sections["faq"] = p.extractSection(content, []string{"FAQ", "Frequently Asked Questions", "常见问题"})
	result.Sections["schema"] = p.extractSection(content, []string{"Schema", "JSON-LD", "Structured Data", "结构化数据"})

	result.ProductMentions = p.extractProductMentions(content)

	return result, nil
}

// ParseSchemaResponse parses a Schema response.
func (p *Parser) ParseSchemaResponse(content string) (map[string]interface{}, error) {
	jsonStr := p.extractJSON(content)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	return schema, nil
}

// ParseAnalysisResponse parses an analysis response.
func (p *Parser) ParseAnalysisResponse(content string) (*ContentAnalysis, error) {
	jsonStr := p.extractJSON(content)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in analysis response")
	}

	var analysis ContentAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis JSON: %w", err)
	}

	return &analysis, nil
}

// extractJSON extracts the JSON portion from content.
func (p *Parser) extractJSON(content string) string {
	// Try to match JSON code block
	re := regexp.MustCompile("```json\\s*([\\s\\S]*?)\\s*```")
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try to match inline code block
	re = regexp.MustCompile("`json\\s*([\\s\\S]*?)\\s*`")
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try to find JSON object directly
	re = regexp.MustCompile("\\{[\\s\\S]*\\}")
	matches = re.FindStringSubmatch(content)
	if len(matches) > 0 {
		jsonStr := strings.TrimSpace(matches[0])
		if json.Valid([]byte(jsonStr)) {
			return jsonStr
		}
	}

	// Try to find JSON array
	re = regexp.MustCompile("\\[[\\s\\S]*\\]")
	matches = re.FindStringSubmatch(content)
	if len(matches) > 0 {
		jsonStr := strings.TrimSpace(matches[0])
		if json.Valid([]byte(jsonStr)) {
			return jsonStr
		}
	}

	return ""
}

// extractSection extracts a specific section from content.
func (p *Parser) extractSection(content string, headings []string) string {
	lines := strings.Split(content, "\n")
	var sectionLines []string
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		for _, heading := range headings {
			if strings.HasPrefix(trimmed, heading) ||
				strings.HasPrefix(trimmed, "# "+heading) ||
				strings.HasPrefix(trimmed, "## "+heading) ||
				strings.HasPrefix(trimmed, "### "+heading) {
				inSection = true
				break
			}
		}

		if inSection && strings.HasPrefix(trimmed, "#") {
			isAnotherSection := false
			for _, heading := range headings {
				if strings.HasPrefix(trimmed, heading) ||
					strings.HasPrefix(trimmed, "# "+heading) ||
					strings.HasPrefix(trimmed, "## "+heading) ||
					strings.HasPrefix(trimmed, "### "+heading) {
					isAnotherSection = true
					break
				}
			}
			if isAnotherSection {
				continue
			}
			break
		}

		if inSection {
			sectionLines = append(sectionLines, line)
		}
	}

	return strings.Join(sectionLines, "\n")
}

// extractProductMentions extracts product mentions from content.
func (p *Parser) extractProductMentions(content string) []string {
	mentions := []string{}

	patterns := []string{
		`Product[：:]\s*([^\n。]+)`,
		`Solution[：:]\s*([^\n。]+)`,
		`Service[：:]\s*([^\n。]+)`,
		`产品[：:]\s*([^\n。]+)`,
		`解决方案[：:]\s*([^\n。]+)`,
		`服务[：:]\s*([^\n。]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				mentions = append(mentions, strings.TrimSpace(match[1]))
			}
		}
	}

	return mentions
}

// ParseResult represents the parsed result from LLM response.
type ParseResult struct {
	OptimizedContent string                 `json:"optimized_content"`
	Sections         map[string]string      `json:"sections"`
	JSONData         string                 `json:"json_data"`
	StructuredData   map[string]interface{} `json:"structured_data"`
	ProductMentions  []string               `json:"product_mentions"`
}

// ExtractMarkdownCodeBlock extracts a markdown code block by language.
func ExtractMarkdownCodeBlock(content string, language string) string {
	pattern := fmt.Sprintf("`{3}%s\\s*([\\s\\S]*?)\\s*`{3}", language)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// CleanMarkdown removes markdown formatting from content.
func CleanMarkdown(content string) string {
	re := regexp.MustCompile("```[\\w]*\\s*")
	content = re.ReplaceAllString(content, "")
	re = regexp.MustCompile("```")
	content = re.ReplaceAllString(content, "")

	re = regexp.MustCompile("`([^`]+)`")
	content = re.ReplaceAllString(content, "$1")

	re = regexp.MustCompile("\\n{3,}")
	content = re.ReplaceAllString(content, "\n\n")

	return strings.TrimSpace(content)
}

// ValidateJSON validates if a string is valid JSON.
func ValidateJSON(jsonStr string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

// ExtractKeyPoints extracts key points from content.
func ExtractKeyPoints(content string) []string {
	points := []string{}

	re := regexp.MustCompile(`(?:^|\n)\s*[-*]\s+([^\n]+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			points = append(points, strings.TrimSpace(match[1]))
		}
	}

	re = regexp.MustCompile(`(?:^|\n)\s*\d+\.\s+([^\n]+)`)
	matches = re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			points = append(points, strings.TrimSpace(match[1]))
		}
	}

	return points
}

// extractSchema extracts Schema markup from content.
func (p *Parser) extractSchema(content string) string {
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
				if json.Valid([]byte(jsonContent)) {
					return fmt.Sprintf(`<script type="application/ld+json">%s</script>`, jsonContent)
				}
			}
		}
	}

	re := regexp.MustCompile(`\{[^"]*"@context"[^}]*\}|\{[^"]*"@type"[^}]*\}`)
	matches := re.FindAllString(content, -1)
	for _, match := range matches {
		if json.Valid([]byte(match)) {
			return fmt.Sprintf(`<script type="application/ld+json">%s</script>`, match)
		}
	}

	return ""
}

// extractFAQ extracts the FAQ section from content.
func (p *Parser) extractFAQ(content string) string {
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
func (p *Parser) extractSummary(content string) string {
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
				return strings.TrimSpace(content[end+2:end+200]) + "..."
			}
		}
	}

	firstParagraphEnd := strings.Index(content, "\n\n")
	if firstParagraphEnd > 0 && firstParagraphEnd < 300 {
		return strings.TrimSpace(content[:firstParagraphEnd])
	}

	if len(content) > 200 {
		return strings.TrimSpace(content[:200]) + "..."
	}

	return content
}

// extractProductMentionsWithProduct extracts product mentions with a specific product name.
func (p *Parser) extractProductMentionsWithProduct(content string, productName string) []string {
	if productName == "" {
		return []string{}
	}

	var mentions []string
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
		mentions = append(mentions, context)

		pos = actualPos + len(productName)
	}

	return mentions
}

// extractDifferentiationPoints extracts differentiation points from content.
func (p *Parser) extractDifferentiationPoints(content string) []string {
	var points []string

	keywords := []string{"unique", "advantage", "different", "compared to", "better than", "独特", "优势", "区别", "不同于", "相比", "优于", "差异"}

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

// extractOptimizations extracts optimization details from content.
func (p *Parser) extractOptimizations(content string) []string {
	var optimizations []string

	patterns := []string{
		`Optimization[：:]\s*([^\n]+)`,
		`Improvement[：:]\s*([^\n]+)`,
		`Enhancement[：:]\s*([^\n]+)`,
		`优化[：:]\s*([^\n]+)`,
		`改进[：:]\s*([^\n]+)`,
		`提升[：:]\s*([^\n]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				optimizations = append(optimizations, strings.TrimSpace(match[1]))
			}
		}
	}

	return optimizations
}

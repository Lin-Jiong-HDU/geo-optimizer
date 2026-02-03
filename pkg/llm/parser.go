package llm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Parser LLM响应解析器
type Parser struct{}

// NewParser 创建解析器
func NewParser() *Parser {
	return &Parser{}
}

// ParseOptimizationResponse 解析优化响应
func (p *Parser) ParseOptimizationResponse(content string) (*ParseResult, error) {
	result := &ParseResult{
		OptimizedContent: content,
		Sections:         make(map[string]string),
	}

	// 尝试提取JSON部分
	if jsonPart := p.extractJSON(content); jsonPart != "" {
		result.JSONData = jsonPart

		// 尝试解析为结构化数据
		var structuredData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonPart), &structuredData); err == nil {
			result.StructuredData = structuredData
		}
	}

	// 提取常见章节
	result.Sections["summary"] = p.extractSection(content, []string{"摘要", "Summary", "概述"})
	result.Sections["faq"] = p.extractSection(content, []string{"FAQ", "常见问题", "常见问答"})
	result.Sections["schema"] = p.extractSection(content, []string{"Schema", "JSON-LD", "结构化数据"})

	// 提取产品提及
	result.ProductMentions = p.extractProductMentions(content)

	return result, nil
}

// ParseSchemaResponse 解析Schema响应
func (p *Parser) ParseSchemaResponse(content string) (map[string]interface{}, error) {
	// 提取JSON部分
	jsonStr := p.extractJSON(content)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	// 解析JSON
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	return schema, nil
}

// ParseAnalysisResponse 解析分析响应
func (p *Parser) ParseAnalysisResponse(content string) (*ContentAnalysis, error) {
	// 提取JSON部分
	jsonStr := p.extractJSON(content)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in analysis response")
	}

	// 解析JSON
	var analysis ContentAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis JSON: %w", err)
	}

	return &analysis, nil
}

// extractJSON 提取JSON部分
func (p *Parser) extractJSON(content string) string {
	// 尝试匹配JSON代码块
	re := regexp.MustCompile("```json\\s*([\\s\\S]*?)\\s*```")
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// 尝试匹配单行代码块
	re = regexp.MustCompile("`json\\s*([\\s\\S]*?)\\s*`")
	matches = re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// 尝试直接查找JSON对象
	re = regexp.MustCompile("\\{[\\s\\S]*\\}")
	matches = re.FindStringSubmatch(content)
	if len(matches) > 0 {
		jsonStr := strings.TrimSpace(matches[0])
		// 验证是否为有效JSON
		if json.Valid([]byte(jsonStr)) {
			return jsonStr
		}
	}

	// 尝试查找JSON数组
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

// extractSection 提取指定章节
func (p *Parser) extractSection(content string, headings []string) string {
	lines := strings.Split(content, "\n")
	var sectionLines []string
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 检查是否为目标章节标题
		for _, heading := range headings {
			if strings.HasPrefix(trimmed, heading) ||
				strings.HasPrefix(trimmed, "# "+heading) ||
				strings.HasPrefix(trimmed, "## "+heading) ||
				strings.HasPrefix(trimmed, "### "+heading) {
				inSection = true
				break
			}
		}

		// 检查是否到达新章节
		if inSection && strings.HasPrefix(trimmed, "#") {
			// 检查是否是另一个目标章节
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
			// 到达其他章节，停止提取
			break
		}

		if inSection {
			sectionLines = append(sectionLines, line)
		}
	}

	return strings.Join(sectionLines, "\n")
}

// extractProductMentions 提取产品提及
func (p *Parser) extractProductMentions(content string) []string {
	// 这个实现可以根据产品关键词进行更复杂的匹配
	// 这里提供一个基础实现
	mentions := []string{}

	// 常见产品提及模式
	patterns := []string{
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

// ParseResult 解析结果
type ParseResult struct {
	OptimizedContent string                 `json:"optimized_content"`
	Sections         map[string]string      `json:"sections"`
	JSONData         string                 `json:"json_data"`
	StructuredData   map[string]interface{} `json:"structured_data"`
	ProductMentions  []string               `json:"product_mentions"`
}

// ExtractMarkdownCodeBlock 提取Markdown代码块
func ExtractMarkdownCodeBlock(content string, language string) string {
	pattern := fmt.Sprintf("`{3}%s\\s*([\\s\\S]*?)\\s*`{3}", language)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// CleanMarkdown 清理Markdown格式
func CleanMarkdown(content string) string {
	// 移除代码块标记
	re := regexp.MustCompile("```[\\w]*\\s*")
	content = re.ReplaceAllString(content, "")
	re = regexp.MustCompile("```")
	content = re.ReplaceAllString(content, "")

	// 移除行内代码标记
	re = regexp.MustCompile("`([^`]+)`")
	content = re.ReplaceAllString(content, "$1")

	// 清理多余空行
	re = regexp.MustCompile("\\n{3,}")
	content = re.ReplaceAllString(content, "\n\n")

	return strings.TrimSpace(content)
}

// ValidateJSON 验证JSON格式
func ValidateJSON(jsonStr string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

// ExtractKeyPoints 提取关键点
func ExtractKeyPoints(content string) []string {
	points := []string{}

	// 查找列表项
	re := regexp.MustCompile(`(?:^|\n)\s*[-*]\s+([^\n]+)`)
	matches := re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			points = append(points, strings.TrimSpace(match[1]))
		}
	}

	// 查找数字列表
	re = regexp.MustCompile(`(?:^|\n)\s*\d+\.\s+([^\n]+)`)
	matches = re.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			points = append(points, strings.TrimSpace(match[1]))
		}
	}

	return points
}

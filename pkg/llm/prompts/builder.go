package prompts

import (
	"fmt"
	"strings"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/config"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// Builder Prompt 构建器
type Builder struct {
	aiProfiles map[string]models.AIPreference
}

// NewBuilder 创建 Prompt 构建器
func NewBuilder() *Builder {
	return &Builder{
		aiProfiles: config.AIProfiles,
	}
}

// SetAIProfiles 设置 AI 平台预设配置
func (b *Builder) SetAIProfiles(profiles map[string]models.AIPreference) {
	b.aiProfiles = profiles
}

// BuildOptimizationPrompt 构建优化 Prompt
func (b *Builder) BuildOptimizationPrompt(req *models.OptimizationRequest) string {
	// 应用 AI 平台预设配置
	if len(req.TargetAI) > 0 && req.AIPreferences == nil {
		req.AIPreferences = make(map[string]models.AIPreference)
		for _, ai := range req.TargetAI {
			if profile, ok := b.aiProfiles[ai]; ok {
				req.AIPreferences[ai] = profile
			}
		}
	}

	// 构建基础 Prompt
	prompt := b.buildBasePrompt(req)

	// 应用 AI 偏好配置
	if len(req.AIPreferences) > 0 {
		preferences := make([]models.AIPreference, 0, len(req.AIPreferences))
		for _, pref := range req.AIPreferences {
			preferences = append(preferences, pref)
		}
		prompt = ApplyAIPreferences(prompt, preferences)
	}

	return prompt
}

// buildBasePrompt 构建基础 Prompt
func (b *Builder) buildBasePrompt(req *models.OptimizationRequest) string {
	var prompt strings.Builder

	prompt.WriteString("请优化以下内容以提高其在AI搜索引擎中的可见性和引用率。\n\n")

	// 原始内容部分
	b.writeSection(&prompt, "原始内容", []string{
		fmt.Sprintf("标题：%s", req.Title),
		fmt.Sprintf("内容：%s", req.Content),
	})

	// 企业信息部分
	b.writeEnterpriseInfo(&prompt, req.Enterprise)

	// 目标 AI 平台
	if len(req.TargetAI) > 0 {
		prompt.WriteString(fmt.Sprintf("【目标AI平台】%v\n\n", req.TargetAI))
	}

	// 关键词
	if len(req.Keywords) > 0 {
		prompt.WriteString(fmt.Sprintf("【关键词】%v\n\n", req.Keywords))
	}

	// 优化策略
	if len(req.Strategies) > 0 {
		prompt.WriteString(fmt.Sprintf("【优化策略】%v\n\n", req.Strategies))
	}

	// 竞品信息
	if len(req.Competitors) > 0 {
		b.writeCompetitorInfo(&prompt, req.Competitors)
	}

	// 输出要求
	b.writeOutputRequirements(&prompt)

	return prompt.String()
}

// writeEnterpriseInfo 写入企业信息
func (b *Builder) writeEnterpriseInfo(prompt *strings.Builder, enterprise models.EnterpriseInfo) {
	prompt.WriteString("【企业信息】\n")
	prompt.WriteString(fmt.Sprintf("公司名称：%s\n", enterprise.CompanyName))
	prompt.WriteString(fmt.Sprintf("产品名称：%s\n", enterprise.ProductName))
	prompt.WriteString(fmt.Sprintf("产品描述：%s\n", enterprise.ProductDescription))

	if len(enterprise.ProductFeatures) > 0 {
		prompt.WriteString(fmt.Sprintf("产品特点：%v\n", enterprise.ProductFeatures))
	}

	if len(enterprise.USP) > 0 {
		prompt.WriteString(fmt.Sprintf("独特卖点：%v\n", enterprise.USP))
	}

	if enterprise.CompanyWebsite != "" {
		prompt.WriteString(fmt.Sprintf("公司网站：%s\n", enterprise.CompanyWebsite))
	}

	if enterprise.ProductURL != "" {
		prompt.WriteString(fmt.Sprintf("产品链接：%s\n", enterprise.ProductURL))
	}

	prompt.WriteString("\n")
}

// writeCompetitorInfo 写入竞品信息
func (b *Builder) writeCompetitorInfo(prompt *strings.Builder, competitors []models.CompetitorInfo) {
	prompt.WriteString("【竞品信息】\n")
	for _, comp := range competitors {
		prompt.WriteString(fmt.Sprintf("- %s: 劣势%v\n", comp.Name, comp.Weaknesses))
		if len(comp.CommonObjections) > 0 {
			prompt.WriteString(fmt.Sprintf("  常见异议：%v\n", comp.CommonObjections))
		}
	}
	prompt.WriteString("\n")
}

// writeOutputRequirements 写入输出要求
func (b *Builder) writeOutputRequirements(prompt *strings.Builder) {
	prompt.WriteString(`请提供优化后的内容，包括：
1. 优化后的完整内容
2. 建议的Schema标记（JSON-LD格式）
3. FAQ部分
4. 内容摘要

确保内容：
- 结构清晰，有明确的标题层级
- 在开头提供关键结论
- 包含权威性引用和数据支持
- 自然地植入企业产品信息
- 针对目标AI平台的偏好进行优化
`)
}

// writeSection 写入一个段落
func (b *Builder) writeSection(prompt *strings.Builder, title string, lines []string) {
	prompt.WriteString(fmt.Sprintf("【%s】\n", title))
	for _, line := range lines {
		prompt.WriteString(line + "\n")
	}
	prompt.WriteString("\n")
}

// BuildStrategyPrompt 构建策略特定的 Prompt
func (b *Builder) BuildStrategyPrompt(strategy models.StrategyType, req *models.OptimizationRequest) string {
	content := req.Content
	extraParams := []string{}

	// 构建企业信息
	enterpriseInfo := b.buildEnterpriseInfo(req)

	// 预处理内容（根据策略）
	switch strategy {
	case models.StrategySchema:
		extraParams = append(extraParams, "Article")
	case models.StrategyFAQ:
		// FAQ需要数量，放在extraParams[0]
		// 企业信息通过enterpriseInfo参数传递
	}

	return BuildStrategyPrompt(strategy, content, enterpriseInfo, extraParams...)
}

// buildEnterpriseInfo 构建企业信息字符串
func (b *Builder) buildEnterpriseInfo(req *models.OptimizationRequest) string {
	if req.Enterprise.ProductName == "" && req.Enterprise.ProductDescription == "" &&
		len(req.Enterprise.Certifications) == 0 && len(req.Enterprise.Awards) == 0 {
		return ""
	}

	var info string
	if req.Enterprise.CompanyName != "" {
		info += fmt.Sprintf("\n【企业信息】\n公司名称：%s", req.Enterprise.CompanyName)
	}
	if req.Enterprise.ProductName != "" {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n产品名称：%s", req.Enterprise.ProductName)
	}
	if req.Enterprise.ProductDescription != "" {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n产品描述：%s", req.Enterprise.ProductDescription)
	}
	if len(req.Enterprise.ProductFeatures) > 0 {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n产品特点：%v", req.Enterprise.ProductFeatures)
	}
	if len(req.Enterprise.USP) > 0 {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n独特卖点：%v", req.Enterprise.USP)
	}
	if len(req.Enterprise.Certifications) > 0 {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n认证信息：%v", req.Enterprise.Certifications)
	}
	if len(req.Enterprise.Awards) > 0 {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n奖项荣誉：%v", req.Enterprise.Awards)
	}
	if len(req.Enterprise.CaseStudies) > 0 {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n案例研究：%v", req.Enterprise.CaseStudies)
	}

	return info
}

// BuildMessages 构建完整的消息列表
func (b *Builder) BuildMessages(req *models.OptimizationRequest) []map[string]string {
	prompt := b.BuildOptimizationPrompt(req)

	return []map[string]string{
		{
			"role":    "system",
			"content": GetSystemPrompt("optimization"),
		},
		{
			"role":    "user",
			"content": prompt,
		},
	}
}

// BuildStrategyMessages 构建策略特定的消息列表
func (b *Builder) BuildStrategyMessages(strategy models.StrategyType, req *models.OptimizationRequest) []map[string]string {
	prompt := b.BuildStrategyPrompt(strategy, req)

	task := "optimization"
	if strategy == models.StrategySchema {
		task = "schema"
	}

	return []map[string]string{
		{
			"role":    "system",
			"content": GetSystemPrompt(task),
		},
		{
			"role":    "user",
			"content": prompt,
		},
	}
}

// BuildAnalysisPrompt 构建分析 Prompt
func (b *Builder) BuildAnalysisPrompt(content string) string {
	return fmt.Sprintf(`请分析以下内容的GEO（生成式搜索引擎优化）质量，并给出评分和建议。

内容：
%s

请以JSON格式返回分析结果，包含以下字段：
- structure_score: 结构化评分（0-100）
- authority_score: 权威性评分（0-100）
- clarity_score: 清晰度评分（0-100）
- citation_score: 可引用性评分（0-100）
- schema_score: Schema完整性评分（0-100）
- total_score: 总分（0-100）
- keywords: 提取的关键词列表
- suggestions: 改进建议列表

只返回JSON，不要包含其他内容。`, content)
}

// BuildSchemaPrompt 构建 Schema 生成 Prompt
func (b *Builder) BuildSchemaPrompt(content string, schemaType string) string {
	if schemaType == "" {
		schemaType = "Article"
	}
	return fmt.Sprintf(StrategyPromptSchema, content, schemaType)
}

// BuildFAQPrompt 构建 FAQ 生成 Prompt
func (b *Builder) BuildFAQPrompt(content string, count int) string {
	if count <= 0 {
		count = 5
	}
	// 企业信息为空，使用模板
	return fmt.Sprintf(StrategyPromptFAQ, count, content, "")
}

// BuildFAQPromptWithEnterprise 构建 FAQ 生成 Prompt（带企业信息）
func (b *Builder) BuildFAQPromptWithEnterprise(content string, count int, enterprise models.EnterpriseInfo) string {
	if count <= 0 {
		count = 5
	}

	// 构建企业信息
	enterpriseInfo := b.buildEnterpriseInfoFromInfo(enterprise)

	return fmt.Sprintf(StrategyPromptFAQ, count, content, enterpriseInfo)
}

// buildEnterpriseInfoFromInfo 从 EnterpriseInfo 构建企业信息字符串
func (b *Builder) buildEnterpriseInfoFromInfo(enterprise models.EnterpriseInfo) string {
	if enterprise.ProductName == "" && enterprise.ProductDescription == "" &&
		len(enterprise.Certifications) == 0 && len(enterprise.Awards) == 0 {
		return ""
	}

	var info string
	if enterprise.CompanyName != "" {
		info += fmt.Sprintf("\n【企业信息】\n公司名称：%s", enterprise.CompanyName)
	}
	if enterprise.ProductName != "" {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n产品名称：%s", enterprise.ProductName)
	}
	if enterprise.ProductDescription != "" {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n产品描述：%s", enterprise.ProductDescription)
	}
	if len(enterprise.ProductFeatures) > 0 {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n产品特点：%v", enterprise.ProductFeatures)
	}
	if len(enterprise.USP) > 0 {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n独特卖点：%v", enterprise.USP)
	}
	if len(enterprise.Certifications) > 0 {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n认证信息：%v", enterprise.Certifications)
	}
	if len(enterprise.Awards) > 0 {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n奖项荣誉：%v", enterprise.Awards)
	}
	if len(enterprise.CaseStudies) > 0 {
		if info == "" {
			info += "\n【企业信息】"
		}
		info += fmt.Sprintf("\n案例研究：%v", enterprise.CaseStudies)
	}

	return info
}

// BuildAuthorityPrompt 构建权威性增强 Prompt
func (b *Builder) BuildAuthorityPrompt(content string, enterprise models.EnterpriseInfo) string {
	// 构建企业信息
	enterpriseInfo := b.buildEnterpriseInfoFromInfo(enterprise)

	return fmt.Sprintf(StrategyPromptAuthority, content, enterpriseInfo)
}

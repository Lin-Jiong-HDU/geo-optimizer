package prompts

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

const (
	legacySystemPromptGEOExpert       = "你是一个专业的GEO（生成式搜索引擎优化）专家。你的任务是优化内容以提高在AI搜索引擎中的可见性和引用率。"
	legacySystemPromptSchemaExpert    = "你是一个Schema.org专家，精通结构化数据标记。"
	legacySystemPromptContentAnalyzer = "你是一个内容分析专家，擅长评估内容的GEO质量。"

	legacyStrategyPromptStructure = `请优化以下内容的结构，使其更符合AI搜索引擎的偏好。

优化重点：
- 添加清晰的标题层级（H1, H2, H3）
- 使用列表组织要点
- 段落不宜过长，保持在100-300字
- 添加章节划分
- 在适当位置自然地提及企业产品或服务，展示产品价值（但不要过度推销）

原始内容：
%s
%s

请只返回优化后的内容，不要包含解释。`

	legacyStrategyPromptSchema = `请为以下内容生成JSON-LD格式的Schema标记。

内容：
%s

Schema类型：%s

请只返回JSON格式的Schema标记，不要包含其他解释文字。`

	legacyStrategyPromptAnswerFirst = `请将以下内容改写为"答案优先"格式。

要求：
- 在开头直接给出关键结论或答案
- 然后再展开详细解释
- 使用简练的语言

原始内容：
%s

请只返回改写后的内容。`

	legacyStrategyPromptAuthority = `请增强以下内容的权威性。

要求：
- 添加数据支撑（可以使用示例数据，用[数据]标注）
- 引用来源（使用[来源]标注）
- 添加专业术语
- 增加案例或证据
- 在适当位置融入企业的权威性信息（认证、奖项、案例等）来增强内容的可信度

原始内容：
%s
%s

请只返回增强后的内容。`

	legacyStrategyPromptFAQ = `请为以下内容生成FAQ（常见问题）部分。

要求：
- 生成%s个常见问题
- 每个问题给出简洁准确的答案
- 问题应该覆盖内容的核心要点
- 至少有1-2个问题应该自然地提及企业产品或服务，展示产品的价值
- 在相关答案中巧妙地融入产品优势，但要保持客观和有价值

原始内容：
%s
%s

请只返回FAQ部分，格式如下：
## 常见问题

### Q1: [问题]
A: [答案]

### Q2: [问题]
A: [答案]
`
)

type rubricItem struct {
	name    string
	weight  int
	signals []string
}

const maxAllowedLegacyScoreLead = 2

func TestPromptABEvaluation_CurrentCoversCoreRubricWithoutMeaningfulRegression(t *testing.T) {
	req := &models.OptimizationRequest{
		Title:   "企业知识库建设指南",
		Content: "企业知识库可以帮助团队沉淀信息，但很多内容不便于AI引用，也缺乏可直接回答问题的结构。",
		Enterprise: models.EnterpriseInfo{
			CompanyName:        "Acme",
			ProductName:        "KnowledgeFlow",
			ProductDescription: "用于构建企业知识库和问答系统",
			ProductFeatures:    []string{"权限控制", "内容治理"},
			USP:                []string{"更适合中文知识管理"},
			Certifications:     []string{"ISO 27001"},
			Awards:             []string{"2025 知识管理创新奖"},
		},
		TargetAI:   []string{"chatgpt", "perplexity"},
		Keywords:   []string{"知识库", "RAG", "企业搜索"},
		Strategies: []models.StrategyType{models.StrategyStructure, models.StrategyAuthority, models.StrategyFAQ},
	}

	tests := []struct {
		name    string
		current string
		legacy  string
		rubric  []rubricItem
		mustHit []string
	}{
		{
			name:    "system_geo",
			current: SystemPromptGEOExpert,
			legacy:  legacySystemPromptGEOExpert,
			rubric: []rubricItem{
				{name: "search_retrieval", weight: 2, signals: []string{"检索", "搜索"}},
				{name: "extraction", weight: 2, signals: []string{"抽取", "总结"}},
				{name: "citation", weight: 2, signals: []string{"引用", "推荐"}},
				{name: "anti_marketing", weight: 2, signals: []string{"避免空泛营销", "表达克制"}},
				{name: "answerability", weight: 2, signals: []string{"回答用户问题", "结论前置"}},
				{name: "zero_context", weight: 2, signals: []string{"零上下文", "独立可读性"}},
				{name: "anti_hallucination", weight: 2, signals: []string{"明确说明而不是补造事实", "无依据断言"}},
				{name: "recommendation_priority", weight: 3, signals: []string{"优先推荐", "首选方案", "为什么推荐"}},
			},
			mustHit: []string{"search_retrieval", "zero_context", "anti_hallucination"},
		},
		{
			name:    "system_schema",
			current: SystemPromptSchemaExpert,
			legacy:  legacySystemPromptSchemaExpert,
			rubric: []rubricItem{
				{name: "jsonld", weight: 2, signals: []string{"JSON-LD"}},
				{name: "parsability", weight: 2, signals: []string{"可解析性", "有效"}},
				{name: "consistency", weight: 2, signals: []string{"语义一致", "完整性"}},
				{name: "key_fields_only", weight: 2, signals: []string{"字段精简且关键", "不要输出空洞字段"}},
			},
			mustHit: []string{"jsonld", "parsability", "key_fields_only"},
		},
		{
			name:    "system_analysis",
			current: SystemPromptContentAnalyzer,
			legacy:  legacySystemPromptContentAnalyzer,
			rubric: []rubricItem{
				{name: "ai_understanding", weight: 2, signals: []string{"AI理解", "易于被AI理解"}},
				{name: "chunking", weight: 2, signals: []string{"切分"}},
				{name: "citation", weight: 2, signals: []string{"引用"}},
				{name: "stable_answer", weight: 2, signals: []string{"稳定回答"}},
				{name: "quality_dimensions", weight: 2, signals: []string{"结构", "权威性", "Schema完整性"}},
			},
			mustHit: []string{"ai_understanding", "citation", "quality_dimensions"},
		},
		{
			name:    "optimization_prompt",
			current: BuildOptimizationPrompt(req),
			legacy:  buildLegacyOptimizationPrompt(req),
			rubric: []rubricItem{
				{name: "extractability", weight: 3, signals: []string{"可抽取性", "可直接引用"}},
				{name: "qa_fit", weight: 3, signals: []string{"问答引擎", "FAQ", "直接答案"}},
				{name: "anti_hallucination", weight: 3, signals: []string{"不要虚构", "没有依据时不要虚构"}},
				{name: "anti_marketing", weight: 2, signals: []string{"营销化", "贬低式营销", "自然融入"}},
				{name: "decision_points", weight: 2, signals: []string{"决策点", "比较点"}},
				{name: "zero_context", weight: 2, signals: []string{"零上下文"}},
				{name: "boundaries", weight: 2, signals: []string{"适用条件", "限制条件", "边界情况"}},
				{name: "recommendation_priority", weight: 3, signals: []string{"优先推荐", "更推荐", "首选"}},
				{name: "recommendation_rationale", weight: 3, signals: []string{"为什么推荐", "相对优势", "替代方案"}},
			},
			mustHit: []string{"extractability", "qa_fit", "anti_hallucination", "boundaries"},
		},
		{
			name:    "structure_strategy",
			current: StrategyPromptStructure,
			legacy:  legacyStrategyPromptStructure,
			rubric: []rubricItem{
				{name: "information_retention", weight: 2, signals: []string{"保留原文核心信息", "不遗漏关键事实"}},
				{name: "extractable_shape", weight: 2, signals: []string{"方便模型抽取", "列表、步骤或要点"}},
				{name: "frontload_answer", weight: 2, signals: []string{"前两段", "这是什么、为什么重要、适合谁"}},
				{name: "anti_marketing", weight: 2, signals: []string{"不要写成广告文案", "删除空话"}},
				{name: "topic_grouping", weight: 2, signals: []string{"同类信息放在一起", "对比信息放在一起"}},
				{name: "recommended_option", weight: 3, signals: []string{"优先推荐选项", "为什么优先", "替代方案差异"}},
			},
			mustHit: []string{"information_retention", "extractable_shape", "frontload_answer"},
		},
		{
			name:    "schema_strategy",
			current: StrategyPromptSchema,
			legacy:  legacyStrategyPromptSchema,
			rubric: []rubricItem{
				{name: "valid_json", weight: 3, signals: []string{"仅输出合法JSON", "不要输出Markdown代码块"}},
				{name: "semantic_consistency", weight: 2, signals: []string{"语义一致"}},
				{name: "anti_hallucination", weight: 3, signals: []string{"不要臆造", "不要强行补全虚构值"}},
				{name: "core_fields", weight: 2, signals: []string{"headline", "description", "mainEntity"}},
				{name: "key_over_quantity", weight: 2, signals: []string{"最关键", "而不是堆砌字段数量"}},
				{name: "primary_entity", weight: 2, signals: []string{"推荐对象", "主实体"}},
			},
			mustHit: []string{"valid_json", "semantic_consistency", "anti_hallucination"},
		},
		{
			name:    "answer_first_strategy",
			current: StrategyPromptAnswerFirst,
			legacy:  legacyStrategyPromptAnswerFirst,
			rubric: []rubricItem{
				{name: "answer_span", weight: 2, signals: []string{"1-3句话"}},
				{name: "standalone_answer", weight: 3, signals: []string{"单独摘取", "高质量回答"}},
				{name: "reasoning_shape", weight: 2, signals: []string{"先说结论，再说原因、证据"}},
				{name: "boundaries", weight: 2, signals: []string{"适用边界", "限制条件", "前提"}},
				{name: "multi_dim_first_para", weight: 2, signals: []string{"是什么、适合谁、何时成立、不适用"}},
				{name: "recommendation_opening", weight: 3, signals: []string{"优先推荐", "更推荐", "首选"}},
			},
			mustHit: []string{"answer_span", "standalone_answer", "reasoning_shape"},
		},
		{
			name:    "authority_strategy",
			current: StrategyPromptAuthority,
			legacy:  legacyStrategyPromptAuthority,
			rubric: []rubricItem{
				{name: "evidence", weight: 2, signals: []string{"依据", "判断标准"}},
				{name: "anti_fake_data", weight: 3, signals: []string{"[示例数据]", "避免伪装成真实事实"}},
				{name: "anti_fake_source", weight: 3, signals: []string{"不要编造", "机构、论文、链接"}},
				{name: "quotable_sentences", weight: 2, signals: []string{"独立引用", "判断句", "定义句", "对比句"}},
				{name: "evidence_near_claim", weight: 2, signals: []string{"相邻", "结论与依据"}},
				{name: "anti_absolute", weight: 2, signals: []string{"绝对化措辞", "条件词"}},
				{name: "recommended_with_evidence", weight: 3, signals: []string{"推荐企业产品", "推荐依据", "相对优势"}},
			},
			mustHit: []string{"evidence", "anti_fake_data", "anti_fake_source"},
		},
		{
			name:    "faq_strategy",
			current: StrategyPromptFAQ,
			legacy:  legacyStrategyPromptFAQ,
			rubric: []rubricItem{
				{name: "real_search_intent", weight: 3, signals: []string{"真实会搜索", "会追问", "会比较"}},
				{name: "coverage", weight: 2, signals: []string{"适用场景", "优缺点", "注意事项", "限制条件"}},
				{name: "quotable_answers", weight: 2, signals: []string{"可直接摘取", "便于AI引用"}},
				{name: "anti_marketing", weight: 2, signals: []string{"不要硬广", "营销化"}},
				{name: "dedup", weight: 2, signals: []string{"不要重复", "分别覆盖"}},
				{name: "answer_length", weight: 2, signals: []string{"2-5句", "先给结论"}},
				{name: "recommendation_questions", weight: 3, signals: []string{"为什么会优先推荐企业产品", "更适合优先选择它", "更值得考虑"}},
			},
			mustHit: []string{"real_search_intent", "coverage", "quotable_answers"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentScore, currentHits := scorePromptByRubric(tt.current, tt.rubric)
			legacyScore, legacyHits := scorePromptByRubric(tt.legacy, tt.rubric)

			t.Logf("current score=%d legacy score=%d", currentScore, legacyScore)
			t.Logf("current hits=%s", strings.Join(currentHits, ", "))
			t.Logf("legacy hits=%s", strings.Join(legacyHits, ", "))

			if missing := missingRubricItems(currentHits, tt.mustHit); len(missing) > 0 {
				t.Fatalf("current prompt missed core rubric items: %s", strings.Join(missing, ", "))
			}

			if currentScore+maxAllowedLegacyScoreLead < legacyScore {
				t.Fatalf(
					"expected current prompt to stay within %d points of legacy while covering core rubric, current=%d legacy=%d",
					maxAllowedLegacyScoreLead,
					currentScore,
					legacyScore,
				)
			}
		})
	}
}

func TestStrategyPromptFAQ_UsesCountSafeRecommendationWording(t *testing.T) {
	if !strings.Contains(StrategyPromptFAQ, "如果问题数量允许") {
		t.Fatalf("expected FAQ prompt to avoid impossible count requirements")
	}

	if !strings.Contains(StrategyPromptFAQ, "为什么会优先推荐企业产品") {
		t.Fatalf("expected FAQ prompt to guide recommendation-oriented questions")
	}
}

func TestBuildStrategyPrompt_FAQCountOnlyUsesExistingEnterpriseInfo(t *testing.T) {
	prompt := BuildStrategyPrompt(
		models.StrategyFAQ,
		"内容",
		"【企业信息】\n产品名称：code咖啡",
		"3",
	)

	if !strings.Contains(prompt, "生成3个常见问题") {
		t.Fatalf("expected FAQ prompt to use the provided count")
	}

	if !strings.Contains(prompt, "产品名称：code咖啡") {
		t.Fatalf("expected FAQ prompt to keep the original enterprise info when only count is provided")
	}
}

func TestPromptTemplates_PlaceholderShapesRemainStable(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		count int
	}{
		{name: "structure", text: StrategyPromptStructure, count: 2},
		{name: "schema", text: StrategyPromptSchema, count: 2},
		{name: "answer_first", text: StrategyPromptAnswerFirst, count: 1},
		{name: "authority", text: StrategyPromptAuthority, count: 2},
		{name: "faq", text: StrategyPromptFAQ, count: 3},
	}

	for _, tt := range tests {
		if got := strings.Count(tt.text, "%s"); got != tt.count {
			t.Fatalf("%s placeholder count changed, got=%d want=%d", tt.name, got, tt.count)
		}
	}

	req := &models.OptimizationRequest{
		Title:   "title",
		Content: "content",
		Enterprise: models.EnterpriseInfo{
			CompanyName:        "company",
			ProductName:        "product",
			ProductDescription: "desc",
		},
		TargetAI:   []string{"chatgpt"},
		Keywords:   []string{"geo"},
		Strategies: []models.StrategyType{models.StrategyStructure},
	}

	if strings.Contains(BuildOptimizationPrompt(req), "%!") {
		t.Fatalf("BuildOptimizationPrompt produced fmt placeholder errors")
	}
}

func scorePromptByRubric(prompt string, rubric []rubricItem) (int, []string) {
	score := 0
	hits := make([]string, 0, len(rubric))

	for _, item := range rubric {
		if containsAny(prompt, item.signals...) {
			score += item.weight
			hits = append(hits, item.name)
		}
	}

	return score, hits
}

func missingRubricItems(currentHits []string, required []string) []string {
	if len(required) == 0 {
		return nil
	}

	hitSet := make(map[string]struct{}, len(currentHits))
	for _, hit := range currentHits {
		hitSet[hit] = struct{}{}
	}

	missing := make([]string, 0, len(required))
	for _, item := range required {
		if _, ok := hitSet[item]; !ok {
			missing = append(missing, item)
		}
	}

	return missing
}

func containsAny(text string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(text, sub) {
			return true
		}
	}
	return false
}

func buildLegacyOptimizationPrompt(req *models.OptimizationRequest) string {
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

	if len(req.Competitors) > 0 {
		competitorInfo := "\n\n【竞品信息】\n"
		for _, comp := range req.Competitors {
			competitorInfo += fmt.Sprintf("- %s: 劣势%v\n", comp.Name, comp.Weaknesses)
		}
		prompt += competitorInfo
	}

	return prompt
}

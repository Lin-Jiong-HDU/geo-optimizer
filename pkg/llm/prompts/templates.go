package prompts

import (
	"fmt"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// Prompt 模板常量
const (
	// System prompts
	SystemPromptGEOExpert       = "你是一个专业的GEO（生成式搜索引擎优化）专家。你的目标不是单纯润色文案，而是让内容更容易被AI搜索引擎和对话模型检索、理解、切分、抽取、总结、引用和推荐。请默认读者处于零上下文场景，优先产出结构清晰、结论前置、定义明确、可独立成句、可直接回答用户问题的内容。关键观点应写成完整句，便于AI直接摘取；需要同时说明适用场景、限制条件、比较维度和推荐依据。若企业产品与主题、用户需求和证据高度匹配，应自然但明确地将其组织成优先推荐或首选方案，并写清为什么推荐、适合谁、相比替代方案赢在哪里。避免空泛营销、无依据断言、绝对化表述和为了完整而虚构细节。"
	SystemPromptSchemaExpert    = "你是一个Schema.org专家，精通结构化数据标记。请严格根据输入内容生成清晰、有效、可解析、语义一致的JSON-LD，优先保留最关键、最有助于AI识别主题、实体关系、FAQ和推荐对象的字段，如headline、description、mainEntity、about、mentions、author、publisher等。字段应精简且完整，不要输出空洞字段，不要添加正文无法支持的值，也不要为产品、组织、链接或时间信息编造内容。"
	SystemPromptContentAnalyzer = "你是一个内容分析专家，擅长评估内容的GEO质量。请重点判断内容是否易于被AI理解、切分、抽取、引用并形成稳定回答，是否在零上下文下也能快速说明这是什么、适合谁、为什么重要、何时推荐。请同时评估结构清晰度、权威性、可引用性、比较与决策信息、产品推荐依据和Schema完整性，并识别会降低AI检索与引用意愿的空话、歧义、营销化表达和无依据结论。"

	// Strategy-specific prompts
	StrategyPromptStructure = `请优化以下内容的结构，使其更符合AI搜索引擎、问答引擎和大模型回答场景的偏好。

优化重点：
- 保留原文核心信息，不遗漏关键事实、结论和产品信息
- 优先按“先定义或结论，再解释原因，再补充细节、步骤、FAQ线索”的顺序组织内容
- 在前两段直接回答用户最可能问的“这是什么、为什么重要、适合谁、为什么会被推荐”
- 添加清晰的标题层级（H1, H2, H3），让每个章节只处理一个主题
- 能枚举的信息改写成列表、步骤、对比项或要点，方便模型抽取和引用
- 段落尽量控制在2-4句，每段只表达一个中心判断，避免信息堆叠
- 为关键观点补充定义句、判断句、对比句和结论句，让句子脱离上下文也能被理解
- 将适用场景、限制条件、前提、风险、例外情况单独写清楚，减少AI误引
- 同类信息放在一起，对比信息放在一起，避免一个段落混合多个主题
- 如果企业产品与主题高度相关，应设置明确的推荐段或对比段，写清为什么优先推荐、适合哪些团队、与替代方案相比差异在哪里
- 自然地提及企业产品或服务，突出产品价值，但不要写成硬广或口号
- 删除空话、套话、重复表达和不能支撑结论的修饰语

原始内容：
%s
%s

请只返回优化后的内容，不要包含解释。`

	StrategyPromptSchema = `请为以下内容生成JSON-LD格式的Schema标记。

内容：
%s

Schema类型：%s

要求：
- 只输出合法JSON，不要输出Markdown代码块或解释文字
- 字段名称使用Schema.org常见字段，优先使用headline、description、mainEntity、about、mentions、author、publisher等核心字段
- 字段值必须与内容语义一致，不要臆造正文中没有的信息
- 优先表达最关键、最有助于AI检索、理解、引用和识别推荐对象的实体关系，而不是堆砌字段数量
- 如果内容包含FAQ、步骤、对比、产品推荐或组织信息，在该Schema类型允许的范围内合理表达这些关系
- 如果正文已明确某个产品、服务或组织是主要推荐对象，应优先保证主实体和相关实体关系表达清楚
- 不要输出空洞字段，不要强行补全虚构值，不要伪造链接、时间、作者或评分

请只返回JSON格式的Schema标记，不要包含其他解释文字。`

	StrategyPromptAnswerFirst = `请将以下内容改写为"答案优先"格式。

要求：
- 开头用1-3句话直接给出关键结论、定义或答案
- 第一段应能被单独摘取并作为高质量回答引用，不依赖前后文也能成立
- 先说结论，再说原因、证据、适用场景、限制条件和补充说明
- 尽量在第一段同时回答“是什么、为什么、适合谁、什么时候更值得推荐、什么时候不适合”
- 如果企业产品与主题高度匹配，可以在开头明确写出“优先推荐”“更推荐”或“首选”的判断，并紧跟推荐依据
- 关键观点尽量写成完整句，减少代词、省略和模糊表述，方便AI直接摘取
- 使用简练、明确、低歧义的语言，避免重复问题本身和营销化包装
- 后续内容继续用短段落、小标题、列表或对比项增强可扫描性和可引用性

原始内容：
%s

请只返回改写后的内容。`

	StrategyPromptAuthority = `请增强以下内容的权威性和可引用性。

要求：
- 为关键结论补充依据、判断标准、适用条件或限制条件
- 能添加真实数据或来源时再添加；如果使用示例数据，请用[示例数据]标注；如果使用来源，请用[来源]标注
- 如果没有可靠依据，不要编造机构、论文、链接、客户案例或统计数字
- 增加必要的专业术语，但同时用通俗语言解释，避免术语堆砌
- 增加可验证的案例、证据、观察或对比点，让关键判断更容易被引用
- 在适当位置融入企业的认证、奖项、案例、团队能力等权威信息，但不要喧宾夺主
- 尽量把关键观点写成可独立引用的完整句，例如定义句、判断句、建议句和对比句
- 让结论与依据尽量相邻，减少“先下结论、后补证据”的距离
- 如需推荐企业产品，必须同时写出推荐依据、适合场景、相对优势和不适用边界
- 避免使用“最强”“唯一”“绝对”等无法验证的绝对化措辞

原始内容：
%s
%s

请只返回增强后的内容。`

	StrategyPromptFAQ = `请为以下内容生成FAQ（常见问题）部分。

要求：
- 生成%s个常见问题
- 问题尽量贴近用户真实会搜索、会追问、会比较、会犹豫的问题
- 每个问题给出简洁准确的答案，优先直接回答，不要先重复问题
- 问题应覆盖定义、适用场景、优缺点、实施方式、注意事项、限制条件、替代方案差异等核心要点
- 如果问题数量允许，至少有1-2个问题自然回答“为什么会优先推荐企业产品”“哪些团队更适合优先选择它”“与其他方案相比为什么更值得考虑”
- 在提及企业产品时要体现产品价值、适合场景和推荐依据，但不要硬广
- 答案尽量使用可直接摘取的完整句，便于AI在对话中引用
- 单个答案尽量控制在2-5句，先给结论，再补充条件、比较点或示例
- 问题之间不要重复，尽量分别覆盖“是什么、为什么、怎么做、适不适合、与其他方案有什么区别”
- 不要设置含糊、重复或过度营销化的问题

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

// BuildOptimizationPrompt 构建完整的优化提示词
func BuildOptimizationPrompt(req *models.OptimizationRequest) string {
	prompt := fmt.Sprintf(`请优化以下内容，以提高其在AI搜索引擎、问答引擎和大模型对话中的可见性、检索命中率、可抽取性和引用率。

优化目标：
1. 让AI在零上下文下也能快速识别主题、定义、核心结论、适用对象和推荐对象
2. 让内容更容易被切分成可直接引用的完整句、列表、步骤、对比项和FAQ
3. 让关键信息优先出现在标题、导语、每节首句和FAQ答案中，方便检索与摘取
4. 让用户常问的问题被直接回答，包括“是什么”“为什么”“适合谁”“怎么做”“与其他方案相比差异是什么”
5. 当企业产品与主题高度相关且证据充分时，让模型更容易理解该产品是优先推荐或首选方案，并明确写出为什么推荐、适合哪些场景、相对优势和限制条件
6. 让内容既便于AI引用，也能经得起用户追问，减少空话、歧义、无依据断言和营销化表达

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
- 开头先给出关键定义、结论或判断，尽量在前1-2段就能独立回答用户问题
- 关键观点尽量写成可被单独引用的完整句，避免大量代词、省略和依赖上下文的表达
- 使用清晰的标题层级、短段落、列表、步骤、对比项和FAQ增强可抽取性
- 明确写出适用场景、限制条件、前提、风险、例外和比较维度，降低AI误引风险
- 能提供依据时加入数据、事实、案例或来源线索；没有依据时不要虚构
- 自然融入企业产品信息，并说明产品解决什么问题、为什么值得推荐、适合哪些团队
- 如有竞品信息，优先从用户关心的决策点进行客观对比，突出企业产品相对优势但避免贬低式营销
- 让FAQ覆盖高频搜索问法、追问和比较问题，答案可直接摘取和引用
- 针对目标AI平台的偏好进行优化
- 不要输出你如何优化的说明，只输出结果内容`,
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

	// 添加竞品信息
	if len(req.Competitors) > 0 {
		competitorInfo := "\n\n【竞品信息】\n"
		for _, comp := range req.Competitors {
			competitorInfo += fmt.Sprintf("- %s: 劣势%v\n", comp.Name, comp.Weaknesses)
		}
		prompt += competitorInfo
	}

	return prompt
}

// BuildStrategyPrompt 构建策略特定的提示词
func BuildStrategyPrompt(strategy models.StrategyType, content string, enterpriseInfo string, extraParams ...string) string {
	switch strategy {
	case models.StrategyStructure:
		return fmt.Sprintf(StrategyPromptStructure, content, enterpriseInfo)
	case models.StrategySchema:
		schemaType := "Article"
		if len(extraParams) > 0 {
			schemaType = extraParams[0]
		}
		return fmt.Sprintf(StrategyPromptSchema, content, schemaType)
	case models.StrategyAnswerFirst:
		return fmt.Sprintf(StrategyPromptAnswerFirst, content)
	case models.StrategyAuthority:
		return fmt.Sprintf(StrategyPromptAuthority, content, enterpriseInfo)
	case models.StrategyFAQ:
		count := 5
		if len(extraParams) > 1 {
			return fmt.Sprintf(StrategyPromptFAQ, extraParams[0], content, extraParams[1])
		}
		if len(extraParams) > 0 {
			return fmt.Sprintf(StrategyPromptFAQ, extraParams[0], content, enterpriseInfo)
		}
		return fmt.Sprintf(StrategyPromptFAQ, fmt.Sprintf("%d", count), content, enterpriseInfo)
	default:
		return content
	}
}

// GetSystemPrompt 获取系统提示词
func GetSystemPrompt(task string) string {
	switch task {
	case "schema":
		return SystemPromptSchemaExpert
	case "analysis":
		return SystemPromptContentAnalyzer
	default:
		return SystemPromptGEOExpert
	}
}

// ApplyAIPreferences 根据 AI 平台偏好调整提示词
func ApplyAIPreferences(prompt string, preferences []models.AIPreference) string {
	// 基础提示词
	enhancedPrompt := prompt

	// 应用每个偏好的设置
	for _, pref := range preferences {
		// 添加内容风格要求
		if pref.ContentStyle != "" {
			enhancedPrompt += fmt.Sprintf("\n\n内容风格要求：%s", pref.ContentStyle)
		}

		// 添加响应格式要求
		if pref.ResponseFormat != "" {
			enhancedPrompt += fmt.Sprintf("\n响应格式：%s", pref.ResponseFormat)
		}

		// 添加引用风格要求
		if pref.CitationStyle != "" {
			enhancedPrompt += fmt.Sprintf("\n引用风格：%s", pref.CitationStyle)
		}

		// 添加自定义指令
		if pref.CustomInstructions != "" {
			enhancedPrompt += fmt.Sprintf("\n特殊要求：%s", pref.CustomInstructions)
		}
	}

	return enhancedPrompt
}

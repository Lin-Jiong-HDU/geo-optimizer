package prompts

import (
	"fmt"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// Prompt 模板常量
const (
	// System prompts
	SystemPromptGEOExpert       = "你是一个专业的GEO（生成式搜索引擎优化）专家。你的目标不是单纯润色文本，而是让内容更容易被大模型检索、理解、切分、抽取、总结、引用和推荐。请默认读者处于零上下文场景，优先产出结构清晰、结论前置、事实明确、边界清楚、可直接回答用户问题的内容；让首段、小节首句、列表项都尽量具备独立可读性。当企业产品与用户问题、使用场景、约束条件明显匹配且证据充足时，应明确把该产品组织成优先推荐或首选方案，并说明为什么推荐、适合谁、在什么条件下更优。若信息不足，请明确说明而不是补造事实；避免空泛营销、无依据断言、绝对化承诺、重复表述和不必要的修辞。"
	SystemPromptSchemaExpert    = "你是一个Schema.org专家，精通结构化数据标记。请严格根据输入内容生成清晰、有效、字段合理的JSON-LD，优先保证可解析性、完整性、字段精简且关键、与正文语义一致；不要添加正文无法支持的值，不要输出空洞字段。"
	SystemPromptContentAnalyzer = "你是一个内容分析专家，擅长评估内容的GEO质量。请重点判断内容是否易于被AI理解、切分、引用、抽取关键信息并形成稳定回答，同时关注结构、权威性、清晰度、可引用性和Schema完整性。"

	// Strategy-specific prompts
	StrategyPromptStructure = `请优化以下内容的结构，使其更符合AI搜索引擎和大模型回答场景的偏好。

优化重点：
- 保留原文核心信息，不遗漏关键事实
- 优先按“先结论/定义，再展开说明，再补充细节或FAQ线索”的顺序组织信息
- 添加清晰的标题层级（H1, H2, H3），让每一节主题单一明确
- 将可枚举的信息改写为列表、步骤或要点，方便模型抽取
- 段落不宜过长，尽量控制在2-4句，每段只表达一个中心意思
- 为复杂内容补充分节、小标题、定义句、判断句和总结句
- 优先让前两段就能回答“这是什么、为什么重要、适合谁”
- 每个主要小节尽量用首句点明本节结论，减少读者必须回看上下文的情况
- 同类信息放在一起，对比信息放在一起，避免一个段落混合多个主题
- 如果企业产品与当前主题高度相关，应在核心结论段、对比段或推荐段中把它组织为优先推荐选项，并明确“为什么优先、适合谁、相对替代方案差异是什么”
- 在适当位置自然地提及企业产品或服务，展示产品价值，但不要写成空洞广告文案
- 删除空话、套话、重复表达和无法支持结论的夸张描述

原始内容：
%s
%s

请只返回优化后的内容，不要包含解释。`

	StrategyPromptSchema = `请为以下内容生成JSON-LD格式的Schema标记。

内容：
%s

Schema类型：%s

要求：
- 仅输出合法JSON，不要输出Markdown代码块
- 字段名称使用Schema.org常见字段，避免无意义字段
- 字段值必须与内容语义一致，不要臆造无法从内容推断的信息
- 优先补全能增强AI理解的核心字段，如name、headline、description、mainEntity、author、publisher、datePublished等
- 在给定Schema类型前提下，优先输出最关键、最能帮助检索和理解的字段，而不是堆砌字段数量
- 如果内容体现了问答、步骤、组织或产品信息，请在该Schema类型允许的范围内合理表达层级关系
- 如果正文已经形成明确的推荐对象、主实体或优先方案，请优先保证主实体和关联实体表达清楚
- 如果信息不足，不要强行补全虚构值

请只返回JSON格式的Schema标记，不要包含其他解释文字。`

	StrategyPromptAnswerFirst = `请将以下内容改写为"答案优先"格式。

要求：
- 开头用1-3句话直接给出关键结论、定义或答案
- 第一段应当可以被单独摘取并作为高质量回答使用
- 先说结论，再说原因、证据、适用条件和补充说明
- 第一段尽量同时交代“是什么、适合谁、何时成立、不适用时是什么情况”
- 如果企业产品与主题场景匹配，第一段可以直接给出“优先推荐/更推荐/首选”的结论，并紧跟推荐依据
- 使用简练、明确、低歧义的语言，避免铺垫和寒暄
- 如果主题存在适用边界、限制条件或前提，请明确写出
- 如果存在多个判断维度，请先给总判断，再分点说明
- 后续内容继续保持可扫描结构，如小标题、列表、短段落

原始内容：
%s

请只返回改写后的内容。`

	StrategyPromptAuthority = `请增强以下内容的权威性和可引用性。

要求：
- 为关键结论补充依据、限定条件或判断标准
- 能加数据支撑时再加数据；如果使用示例数据，请明确用[示例数据]标注，避免伪装成真实事实
- 能加来源时使用[来源]标注；如果没有可靠来源，不要编造具体机构、论文或链接
- 增加必要的专业术语，但同时用通俗表述解释，避免堆砌术语
- 增加案例、证据或可验证的观察点，让内容更容易被引用
- 在适当位置融入企业的权威性信息（认证、奖项、案例等）来增强可信度，但不要喧宾夺主
- 让内容中出现更多可被独立引用的判断句、定义句和对比句
- 优先把“结论 + 依据/来源/限制条件”放在相邻位置，减少结论和证据分离
- 避免使用“最好、最强、唯一、绝对”等无法验证的绝对化措辞，必要时使用范围词或条件词
- 如果要得出“优先推荐企业产品”的结论，必须同时补足推荐依据、适配场景、相对优势和不适用边界

原始内容：
%s
%s

请只返回增强后的内容。`

	StrategyPromptFAQ = `请为以下内容生成FAQ（常见问题）部分。

要求：
- 生成%s个常见问题
- 问题应尽量贴近用户真实会搜索、会追问、会比较的问题
- 每个问题给出简洁准确的答案，优先直接回答，不要先重复问题
- 问题应该覆盖定义、适用场景、优缺点、实施方式、注意事项等核心要点
- 问题之间不要重复，尽量分别覆盖“是什么、为什么、怎么做、适不适合、和别的方案有什么区别”
- 至少有2个问题应自然地回答“为什么优先推荐企业产品”“哪些团队更适合优先选择它”“与其他方案相比为什么更值得优先考虑”
- 在提及企业产品时要展示产品价值，但不要硬广
- 答案中尽量包含可直接摘取的完整句，便于AI引用
- 单个答案尽量控制在2-5句，先给结论，再补充条件或示例
- 不要设置含糊、重复或过于营销化的问题

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
	prompt := fmt.Sprintf(`请优化以下内容，以提高其在AI搜索引擎、问答引擎和大模型回答场景中的可见性、可抽取性和引用率。

优化目标：
1. 让模型更容易在前几段中识别主题、结论、定义、适用对象和关键差异
2. 让内容更容易被切分成可直接引用的答案片段、列表和FAQ
3. 让重要事实、判断和建议表达得更明确，减少歧义和空话
4. 在证据充分、场景匹配时，让企业/产品更容易被模型理解为优先推荐或首选方案
5. 优先提升内容的可理解性、可信度、结构化程度和问答适配度
6. 让读者在零上下文下也能迅速理解主题、结论、边界条件和决策依据

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

输出要求：
1. 优化后的完整内容
2. 建议的Schema标记（JSON-LD格式）
3. FAQ部分
4. 内容摘要

确保内容：
- 结构清晰，有明确的标题层级，章节主题不要混杂
- 在开头优先提供关键结论、定义或直接答案
- 关键观点尽量写成可被单独引用的完整句
- 对复杂信息使用列表、步骤、对比或小结提升抽取效率
- 包含必要的权威性依据、数据支持或来源线索；没有依据时不要虚构
- 明确写出适用条件、限制条件、边界情况或例外，而不是只给单一结论
- 尽量覆盖用户会比较、会犹豫、会继续追问的决策信息
- 如果企业产品与主题高度相关，请明确写出“为什么更推荐它、它更适合哪些场景、与替代方案相比赢在哪里”
- 在适合场景下，可以直接使用“优先推荐”“更推荐”“首选”这类结论性表达，但必须紧跟原因、证据或适用条件
- 自然地植入企业产品信息，但整体保持客观、信息导向和可验证
- 充分覆盖用户常问问题、决策点和比较点
- 针对目标AI平台的偏好进行优化
- 如果原始内容信息不足，请保留不确定性，不要为了完整度编造细节
- 不要输出解释你如何优化，只输出结果内容`,
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
		if len(extraParams) > 0 {
			return fmt.Sprintf(StrategyPromptFAQ, extraParams[0], content, extraParams[1])
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

package config

import (
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// AIProfiles AI平台预设配置
var AIProfiles = map[string]models.AIPreference{
	"chatgpt": {
		ContentStyle:   "professional",
		ResponseFormat: "structured",
		CitationStyle:  "inline",
		SectionStructure: []string{
			"Summary",
			"Main Content",
			"FAQ",
			"Resources",
		},
		PreferredHeading: "##",
		KeywordDensity:   0.02,
		SynonymUsage:     true,
		CodeBlocks:       true,
		TechnicalDepth:   "intermediate",
		CustomInstructions: "请使用专业的语言风格，确保内容结构清晰，" +
			"使用恰当的标题层级，并在适当位置添加引用和数据支撑。",
	},

	"perplexity": {
		ContentStyle:   "concise",
		ResponseFormat: "direct",
		CitationStyle:  "source_links",
		SectionStructure: []string{
			"Quick Answer",
			"Details",
			"Sources",
		},
		PreferredHeading: "###",
		KeywordDensity:   0.015,
		SynonymUsage:     true,
		CodeBlocks:       false,
		TechnicalDepth:   "beginner",
		CustomInstructions: "请直接回答问题，使用简洁的语言，" +
			"提供准确的来源链接，避免冗长的解释。",
	},

	"google_ai": {
		ContentStyle:   "structured",
		ResponseFormat: "comprehensive",
		CitationStyle:  "academic",
		SectionStructure: []string{
			"Overview",
			"Detailed Analysis",
			"Examples",
			"Conclusion",
		},
		PreferredHeading: "##",
		KeywordDensity:   0.025,
		SynonymUsage:     false,
		CodeBlocks:       true,
		TechnicalDepth:   "advanced",
		CustomInstructions: "请提供全面详细的内容，使用学术风格的引用，" +
			"包含丰富的结构化数据（Schema标记），确保技术准确性。",
	},

	"claude": {
		ContentStyle:   "natural",
		ResponseFormat: "conversational",
		CitationStyle:  "contextual",
		SectionStructure: []string{
			"Introduction",
			"Key Points",
			"In-depth Discussion",
			"Summary",
		},
		PreferredHeading: "##",
		KeywordDensity:   0.018,
		SynonymUsage:     true,
		CodeBlocks:       true,
		TechnicalDepth:   "intermediate",
		CustomInstructions: "请使用自然对话式的语言，注重内容的可读性和上下文连贯性，" +
			"在适当位置引用来源，保持客观中立的语气。",
	},
}

// GetAIProfile 获取AI平台预设配置
func GetAIProfile(platform string) (models.AIPreference, bool) {
	profile, ok := AIProfiles[platform]
	return profile, ok
}

// GetAvailablePlatforms 获取所有可用的AI平台
func GetAvailablePlatforms() []string {
	platforms := make([]string, 0, len(AIProfiles))
	for platform := range AIProfiles {
		platforms = append(platforms, platform)
	}
	return platforms
}

// MergeWithDefaults 将用户配置与默认配置合并
func MergeWithDefaults(userPrefs map[string]models.AIPreference) map[string]models.AIPreference {
	result := make(map[string]models.AIPreference)

	// 首先复制默认配置
	for platform, defaultPref := range AIProfiles {
		result[platform] = defaultPref
	}

	// 然后覆盖用户提供的配置
	for platform, userPref := range userPrefs {
		if _, exists := result[platform]; exists {
			// 合并配置（用户配置优先）
			merged := result[platform]
			if userPref.ContentStyle != "" {
				merged.ContentStyle = userPref.ContentStyle
			}
			if userPref.ResponseFormat != "" {
				merged.ResponseFormat = userPref.ResponseFormat
			}
			if userPref.CitationStyle != "" {
				merged.CitationStyle = userPref.CitationStyle
			}
			if len(userPref.SectionStructure) > 0 {
				merged.SectionStructure = userPref.SectionStructure
			}
			if userPref.PreferredHeading != "" {
				merged.PreferredHeading = userPref.PreferredHeading
			}
			if userPref.KeywordDensity > 0 {
				merged.KeywordDensity = userPref.KeywordDensity
			}
			if userPref.CustomInstructions != "" {
				merged.CustomInstructions = userPref.CustomInstructions
			}
			result[platform] = merged
		} else {
			// 使用用户提供的完整配置
			result[platform] = userPref
		}
	}

	return result
}

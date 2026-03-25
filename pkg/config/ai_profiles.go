package config

import (
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// AIProfiles contains preset configurations for AI platforms.
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
		CustomInstructions: "Use professional language style, ensure clear content structure, " +
			"use appropriate heading levels, and add citations and data support where appropriate.",
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
		CustomInstructions: "Answer directly, use concise language, " +
			"provide accurate source links, avoid lengthy explanations.",
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
		CustomInstructions: "Provide comprehensive and detailed content, use academic citation style, " +
			"include rich structured data (Schema markup), ensure technical accuracy.",
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
		CustomInstructions: "Use natural conversational language, focus on readability and contextual coherence, " +
			"cite sources where appropriate, maintain an objective and neutral tone.",
	},
}

// GetAIProfile returns the preset configuration for an AI platform.
func GetAIProfile(platform string) (models.AIPreference, bool) {
	profile, ok := AIProfiles[platform]
	return profile, ok
}

// GetAvailablePlatforms returns all available AI platforms.
func GetAvailablePlatforms() []string {
	platforms := make([]string, 0, len(AIProfiles))
	for platform := range AIProfiles {
		platforms = append(platforms, platform)
	}
	return platforms
}

// MergeWithDefaults merges user preferences with default configurations.
func MergeWithDefaults(userPrefs map[string]models.AIPreference) map[string]models.AIPreference {
	result := make(map[string]models.AIPreference)

	for platform, defaultPref := range AIProfiles {
		result[platform] = defaultPref
	}

	for platform, userPref := range userPrefs {
		if _, exists := result[platform]; exists {
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
			result[platform] = userPref
		}
	}

	return result
}

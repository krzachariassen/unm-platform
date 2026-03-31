package handler

import (
	"strings"
	"unicode/utf8"
)

// ComplexityTier represents the estimated cognitive complexity of a user question.
type ComplexityTier string

const (
	TierSimple  ComplexityTier = "simple"
	TierMedium  ComplexityTier = "medium"
	TierComplex ComplexityTier = "complex"
)

// complexitySignals are weighted keywords/phrases that increase the complexity score.
var complexSignals = []struct {
	pattern string
	weight  int
}{
	// Strategic / multi-step planning
	{"restructur", 4},
	{"reorganiz", 4},
	{"scaling the org", 4},
	{"scale the org", 4},
	{"distribute hc", 3},
	{"distribute headcount", 3},
	{"hire .* engineer", 3},
	{"plan to hire", 3},
	{"comprehensive", 3},
	{"deeply analy", 3},
	{"deep analysis", 3},
	{"full analysis", 3},

	// Multi-entity scope
	{"all teams", 2},
	{"each team", 2},
	{"every team", 2},
	{"across the org", 2},
	{"entire system", 2},
	{"across all", 2},

	// Comparative / exploratory
	{"compare", 1},
	{"trade-off", 1},
	{"tradeoff", 1},
	{"pros and cons", 1},
	{"not just", 2},
	{"but also", 1},
	{"in terms of", 1},

	// Structural change verbs (multi-action)
	{"split.*merge", 3},
	{"merge.*split", 3},
	{"redistribution", 2},
	{"realign", 2},
	{"alignment to", 2},

	// Analytical depth
	{"analyz", 2},
	{"assess", 1},
	{"evaluat", 1},
	{"recommend", 2},
	{"suggest", 1},
	{"consolidat", 2},
	{"prioriti", 1},
	{"roadmap", 2},
	{"action plan", 2},
	{"strategy", 1},
	{"strategic", 2},
	{"overloaded", 1},
	{"cognitive load", 1},
	{"reduce", 1},
}

// simplicitySignals are patterns that indicate a simple, direct question.
var simplicitySignals = []string{
	"what happens if",
	"which team owns",
	"who owns",
	"how many",
	"what is the",
	"what are the",
	"does team",
	"is service",
	"list all",
	"show me",
	"what does",
	"where is",
}

// ClassifyComplexity scores a user question and returns a complexity tier.
// The heuristic combines question length, strategic keyword density, and
// simplicity pattern matching. No AI call is needed — this runs in <1ms.
func ClassifyComplexity(question string) ComplexityTier {
	lower := strings.ToLower(question)
	charCount := utf8.RuneCountInString(question)
	wordCount := len(strings.Fields(question))

	// Check simplicity patterns first
	for _, pat := range simplicitySignals {
		if strings.Contains(lower, pat) && charCount < 120 {
			return TierSimple
		}
	}

	// Very short questions are simple
	if charCount < 60 {
		return TierSimple
	}

	// Score complexity signals
	score := 0
	for _, sig := range complexSignals {
		if strings.Contains(lower, sig.pattern) {
			score += sig.weight
		}
	}

	// Length contributes to complexity
	if charCount > 200 {
		score += 1
	}
	if charCount > 400 {
		score += 2
	}
	if wordCount > 50 {
		score += 1
	}

	// Multiple sentences suggest multi-part questions
	sentences := strings.Count(question, ".") + strings.Count(question, "?") + strings.Count(question, "!")
	if sentences >= 3 {
		score += 1
	}

	if score >= 5 {
		return TierComplex
	}
	if score >= 2 {
		return TierMedium
	}
	return TierSimple
}

// TierConfigKey returns the config lookup key for the given tier.
// This is used with ModelForCategory, TimeoutForCategory, and reasoningEffortForCategory.
func TierConfigKey(tier ComplexityTier) string {
	switch tier {
	case TierSimple:
		return "advisor/simple"
	case TierComplex:
		return "advisor/complex"
	default:
		return "advisor/general"
	}
}

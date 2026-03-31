package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassifyComplexity_SimpleQuestions(t *testing.T) {
	cases := []string{
		"What happens if I move service-a to team-b?",
		"Which team owns the indexer service?",
		"How many services does inca-core have?",
		"List all capabilities",
		"What is the cognitive load of team X?",
		"Show me the bottlenecks",
	}
	for _, q := range cases {
		assert.Equal(t, TierSimple, ClassifyComplexity(q), "expected simple: %s", q)
	}
}

func TestClassifyComplexity_MediumQuestions(t *testing.T) {
	cases := []string{
		"Which teams are overloaded and what changes would you recommend to reduce their cognitive load?",
		"Analyze the fragmented capabilities and suggest how to consolidate them across teams",
		"Should we split team inca-core-dev? Evaluate the trade-offs of splitting versus keeping it as is",
	}
	for _, q := range cases {
		tier := ClassifyComplexity(q)
		assert.NotEqual(t, TierSimple, tier, "expected medium or complex (not simple): %s", q)
	}
}

func TestClassifyComplexity_ComplexQuestions(t *testing.T) {
	cases := []string{
		"We plan to hire 10 engineers, analyze each team and find the best way to distribute HC. Not just as-is but deeply analyse how each team can scale in terms of split, merge, redistribution and alignment to Team Topologies.",
		"Do a comprehensive restructuring analysis of the entire system. Evaluate every team's cognitive load, recommend service redistribution, and provide a prioritized action roadmap with trade-offs for each recommendation.",
		"Analyze all teams and recommend a full reorganization strategy. Consider splitting overloaded teams, merging underutilized ones, redistributing services for better alignment, and creating a roadmap.",
	}
	for _, q := range cases {
		assert.Equal(t, TierComplex, ClassifyComplexity(q), "expected complex: %s", q)
	}
}

func TestTierConfigKey(t *testing.T) {
	assert.Equal(t, "advisor/simple", TierConfigKey(TierSimple))
	assert.Equal(t, "advisor/general", TierConfigKey(TierMedium))
	assert.Equal(t, "advisor/complex", TierConfigKey(TierComplex))
}

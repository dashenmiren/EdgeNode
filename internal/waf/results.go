package waf

// PerformResult action performing result
type PerformResult struct {
	ContinueRequest bool
	GoNextGroup     bool
	GoNextSet       bool
	IsAllowed       bool
	AllowScope      AllowScope
}

// MatchResult request match result
type MatchResult struct {
	GoNext         bool
	HasRequestBody bool
	Group          *RuleGroup
	Set            *RuleSet
	IsAllowed      bool
	AllowScope     AllowScope
}
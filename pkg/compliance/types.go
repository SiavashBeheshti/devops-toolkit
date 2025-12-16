package compliance

import "time"

// CheckStatus represents the status of a compliance check
type CheckStatus string

const (
	StatusPassed  CheckStatus = "passed"
	StatusFailed  CheckStatus = "failed"
	StatusSkipped CheckStatus = "skipped"
	StatusWarning CheckStatus = "warning"
)

// CheckResult represents the result of a compliance check
type CheckResult struct {
	RuleID      string      `json:"rule_id"`
	RuleName    string      `json:"rule_name"`
	Category    string      `json:"category"`
	Severity    string      `json:"severity"`
	Status      CheckStatus `json:"status"`
	Resource    string      `json:"resource"`
	Message     string      `json:"message"`
	Remediation string      `json:"remediation,omitempty"`
}

// CheckOptions contains options for compliance checks
type CheckOptions struct {
	Namespace   string
	Image       string
	Path        string
	SkipRules   []string
	OnlyRules   []string
	MinSeverity string
}

// Policy represents a compliance policy
type Policy struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Category    string `yaml:"category" json:"category"`
	Severity    string `yaml:"severity" json:"severity"`
	Description string `yaml:"description" json:"description"`
	Remediation string `yaml:"remediation" json:"remediation"`
}

// Report represents a compliance report
type Report struct {
	Title       string        `json:"title"`
	GeneratedAt time.Time     `json:"generated_at"`
	Summary     ReportSummary `json:"summary"`
	Results     []CheckResult `json:"results"`
}

// ReportSummary contains report summary statistics
type ReportSummary struct {
	Total   int     `json:"total"`
	Passed  int     `json:"passed"`
	Failed  int     `json:"failed"`
	Skipped int     `json:"skipped"`
	Score   float64 `json:"score"`
}


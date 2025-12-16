package compliance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/beheshti/devops-toolkit/pkg/compliance"
	"github.com/beheshti/devops-toolkit/pkg/output"
	"github.com/spf13/cobra"
)

func newReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate compliance report",
		Long: `Generate a comprehensive compliance report.

Output formats:
  table     Console table output (default)
  json      JSON format for programmatic use
  junit     JUnit XML format for CI integration
  html      HTML report for sharing`,
		RunE: runReport,
	}

	cmd.Flags().StringP("format", "f", "table", "Output format (table, json, junit, html)")
	cmd.Flags().StringP("file", "o", "", "Output file path")
	cmd.Flags().String("title", "Compliance Report", "Report title")
	cmd.Flags().Bool("include-passed", true, "Include passed checks in report")

	return cmd
}

func runReport(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	outputFile, _ := cmd.Flags().GetString("file")
	title, _ := cmd.Flags().GetString("title")
	includePassed, _ := cmd.Flags().GetBool("include-passed")

	output.StartSpinner("Running compliance checks...")

	// Run all checks
	opts := compliance.CheckOptions{}
	results, err := runAllChecks(context.Background(), opts)
	if err != nil {
		output.SpinnerError("Failed to run checks")
		return err
	}

	output.SpinnerSuccess(fmt.Sprintf("Completed %d checks", len(results)))

	// Filter results
	if !includePassed {
		var filtered []compliance.CheckResult
		for _, r := range results {
			if r.Status != compliance.StatusPassed {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	// Generate report
	report := compliance.Report{
		Title:       title,
		GeneratedAt: time.Now(),
		Results:     results,
	}

	// Calculate summary
	for _, r := range results {
		switch r.Status {
		case compliance.StatusPassed:
			report.Summary.Passed++
		case compliance.StatusFailed:
			report.Summary.Failed++
		case compliance.StatusSkipped:
			report.Summary.Skipped++
		}
	}
	report.Summary.Total = len(results)
	if report.Summary.Total > 0 {
		report.Summary.Score = float64(report.Summary.Passed) / float64(report.Summary.Total-report.Summary.Skipped) * 100
	}

	// Output based on format
	var reportOutput string

	switch format {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		reportOutput = string(data)
	case "junit":
		reportOutput = generateJUnitReport(report)
	case "html":
		reportOutput = generateHTMLReport(report)
	default: // table
		displayResults(results)
		return nil
	}

	// Write to file or stdout
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(reportOutput), 0644)
		if err != nil {
			return fmt.Errorf("failed to write report: %w", err)
		}
		output.Successf("Report written to %s", outputFile)
	} else {
		fmt.Println(reportOutput)
	}

	return nil
}

func generateJUnitReport(report compliance.Report) string {
	// JUnit XML format for CI integration
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="Compliance Checks" tests="%d" failures="%d" time="0">
`
	xml = fmt.Sprintf(xml, report.Summary.Total, report.Summary.Failed)

	// Group by category as test suites
	byCategory := make(map[string][]compliance.CheckResult)
	for _, r := range report.Results {
		byCategory[r.Category] = append(byCategory[r.Category], r)
	}

	for category, results := range byCategory {
		failures := 0
		for _, r := range results {
			if r.Status == compliance.StatusFailed {
				failures++
			}
		}

		xml += fmt.Sprintf(`  <testsuite name="%s" tests="%d" failures="%d">
`, category, len(results), failures)

		for _, r := range results {
			xml += fmt.Sprintf(`    <testcase name="%s" classname="%s">
`, r.RuleID, r.Resource)

			if r.Status == compliance.StatusFailed {
				xml += fmt.Sprintf(`      <failure message="%s" type="%s">%s</failure>
`, r.Message, r.Severity, r.Message)
			} else if r.Status == compliance.StatusSkipped {
				xml += `      <skipped/>
`
			}

			xml += `    </testcase>
`
		}

		xml += `  </testsuite>
`
	}

	xml += `</testsuites>`
	return xml
}

func generateHTMLReport(report compliance.Report) string {
	// Generate a clean HTML report
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0f172a; color: #e2e8f0; line-height: 1.6; }
        .container { max-width: 1200px; margin: 0 auto; padding: 2rem; }
        h1 { color: #7c3aed; margin-bottom: 0.5rem; }
        .subtitle { color: #64748b; margin-bottom: 2rem; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 1rem; margin-bottom: 2rem; }
        .stat { background: #1e293b; padding: 1.5rem; border-radius: 8px; text-align: center; }
        .stat-value { font-size: 2rem; font-weight: bold; }
        .stat-label { color: #64748b; font-size: 0.875rem; }
        .passed { color: #10b981; }
        .failed { color: #ef4444; }
        .warning { color: #f59e0b; }
        .score-bar { height: 8px; background: #374151; border-radius: 4px; overflow: hidden; margin-top: 1rem; }
        .score-fill { height: 100%%; background: linear-gradient(90deg, #10b981, #7c3aed); }
        .category { background: #1e293b; border-radius: 8px; margin-bottom: 1rem; overflow: hidden; }
        .category-header { padding: 1rem; background: #334155; font-weight: bold; }
        table { width: 100%%; border-collapse: collapse; }
        th, td { padding: 0.75rem 1rem; text-align: left; border-bottom: 1px solid #374151; }
        th { background: #1e293b; color: #94a3b8; font-weight: 500; }
        .badge { display: inline-block; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; font-weight: bold; }
        .badge-critical { background: #ef4444; }
        .badge-high { background: #f97316; }
        .badge-medium { background: #f59e0b; color: #000; }
        .badge-low { background: #06b6d4; }
        .status-icon { width: 20px; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <h1>%s</h1>
        <p class="subtitle">Generated: %s</p>
        
        <div class="summary">
            <div class="stat">
                <div class="stat-value">%d</div>
                <div class="stat-label">Total Checks</div>
            </div>
            <div class="stat">
                <div class="stat-value passed">%d</div>
                <div class="stat-label">Passed</div>
            </div>
            <div class="stat">
                <div class="stat-value failed">%d</div>
                <div class="stat-label">Failed</div>
            </div>
            <div class="stat">
                <div class="stat-value">%.1f%%</div>
                <div class="stat-label">Score</div>
                <div class="score-bar"><div class="score-fill" style="width: %.1f%%"></div></div>
            </div>
        </div>
`

	html = fmt.Sprintf(html,
		report.Title,
		report.Title,
		report.GeneratedAt.Format("2006-01-02 15:04:05"),
		report.Summary.Total,
		report.Summary.Passed,
		report.Summary.Failed,
		report.Summary.Score,
		report.Summary.Score,
	)

	// Group by category
	byCategory := make(map[string][]compliance.CheckResult)
	for _, r := range report.Results {
		byCategory[r.Category] = append(byCategory[r.Category], r)
	}

	for category, results := range byCategory {
		html += fmt.Sprintf(`
        <div class="category">
            <div class="category-header">%s</div>
            <table>
                <thead>
                    <tr>
                        <th class="status-icon">Status</th>
                        <th>Severity</th>
                        <th>Rule</th>
                        <th>Resource</th>
                        <th>Message</th>
                    </tr>
                </thead>
                <tbody>
`, category)

		for _, r := range results {
			statusIcon := "✓"
			statusClass := "passed"
			if r.Status == compliance.StatusFailed {
				statusIcon = "✗"
				statusClass = "failed"
			} else if r.Status == compliance.StatusSkipped {
				statusIcon = "○"
				statusClass = ""
			}

			severityClass := "low"
			switch r.Severity {
			case "critical":
				severityClass = "critical"
			case "high":
				severityClass = "high"
			case "medium":
				severityClass = "medium"
			}

			html += fmt.Sprintf(`
                    <tr>
                        <td class="status-icon %s">%s</td>
                        <td><span class="badge badge-%s">%s</span></td>
                        <td>%s</td>
                        <td>%s</td>
                        <td>%s</td>
                    </tr>
`, statusClass, statusIcon, severityClass, r.Severity, r.RuleID, r.Resource, r.Message)
		}

		html += `
                </tbody>
            </table>
        </div>
`
	}

	html += `
    </div>
</body>
</html>`

	return html
}


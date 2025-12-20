package compliance

import (
	"context"
	"fmt"
	"strings"

	"github.com/beheshti/devops-toolkit/pkg/compliance"
	"github.com/beheshti/devops-toolkit/pkg/completion"
	"github.com/beheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [target]",
		Short: "Run compliance checks",
		Long: `Run compliance checks against various targets.

Targets:
  k8s           Check Kubernetes resources
  docker        Check Docker images and containers
  files         Check configuration files
  all           Run all available checks

Examples:
  devops-toolkit compliance check k8s
  devops-toolkit compliance check docker --image nginx:latest
  devops-toolkit compliance check files --path ./manifests`,
		Args:              cobra.MinimumNArgs(1),
		RunE:              runCheck,
		SilenceUsage:      true, // Don't show usage on compliance failures
		ValidArgsFunction: completion.ComplianceTargetCompletion,
	}

	cmd.Flags().String("image", "", "Docker image to check")
	cmd.Flags().String("path", ".", "Path to files to check")
	cmd.Flags().StringP("namespace", "n", "", "Kubernetes namespace")
	cmd.Flags().StringSlice("skip", nil, "Rules to skip")
	cmd.Flags().StringSlice("only", nil, "Only run these rules")
	cmd.Flags().String("severity", "", "Minimum severity to report (low, medium, high, critical)")
	cmd.Flags().Bool("fail-on-warn", false, "Exit with error on warnings")

	// Register flag completions
	_ = cmd.RegisterFlagCompletionFunc("namespace", completion.NamespaceCompletion)
	_ = cmd.RegisterFlagCompletionFunc("image", completion.ImageCompletion)
	_ = cmd.RegisterFlagCompletionFunc("severity", completion.SeverityCompletion)

	return cmd
}

func runCheck(cmd *cobra.Command, args []string) error {
	target := strings.ToLower(args[0])

	output.Header("Compliance Check")

	skipRules, _ := cmd.Flags().GetStringSlice("skip")
	onlyRules, _ := cmd.Flags().GetStringSlice("only")
	minSeverity, _ := cmd.Flags().GetString("severity")

	opts := compliance.CheckOptions{
		SkipRules:   skipRules,
		OnlyRules:   onlyRules,
		MinSeverity: minSeverity,
	}

	var results []compliance.CheckResult
	var err error

	switch target {
	case "k8s", "kubernetes":
		namespace, _ := cmd.Flags().GetString("namespace")
		opts.Namespace = namespace
		output.StartSpinner("Checking Kubernetes resources...")
		results, err = runK8sChecks(cmd.Context(), opts)
	case "docker":
		imageName, _ := cmd.Flags().GetString("image")
		opts.Image = imageName
		output.StartSpinner("Checking Docker resources...")
		results, err = runDockerChecks(cmd.Context(), opts)
	case "files", "file":
		path, _ := cmd.Flags().GetString("path")
		opts.Path = path
		output.StartSpinner("Checking configuration files...")
		results, err = runFileChecks(cmd.Context(), opts)
	case "all":
		output.StartSpinner("Running all compliance checks...")
		results, err = runAllChecks(cmd.Context(), opts)
	default:
		return fmt.Errorf("unknown target: %s", target)
	}

	if err != nil {
		output.SpinnerError("Check failed")
		return err
	}

	output.StopSpinner()
	displayResults(results)

	// Determine exit status
	failOnWarn, _ := cmd.Flags().GetBool("fail-on-warn")
	hasErrors := false
	hasWarnings := false

	for _, r := range results {
		if r.Status == compliance.StatusFailed {
			if r.Severity == "critical" || r.Severity == "high" {
				hasErrors = true
			} else {
				hasWarnings = true
			}
		}
	}

	if hasErrors || (failOnWarn && hasWarnings) {
		return fmt.Errorf("compliance check failed")
	}

	return nil
}

func runK8sChecks(ctx context.Context, opts compliance.CheckOptions) ([]compliance.CheckResult, error) {
	checker := compliance.NewK8sChecker(opts)
	return checker.Run(ctx)
}

func runDockerChecks(ctx context.Context, opts compliance.CheckOptions) ([]compliance.CheckResult, error) {
	checker := compliance.NewDockerChecker(opts)
	return checker.Run(ctx)
}

func runFileChecks(ctx context.Context, opts compliance.CheckOptions) ([]compliance.CheckResult, error) {
	checker := compliance.NewFileChecker(opts)
	return checker.Run(ctx)
}

func runAllChecks(ctx context.Context, opts compliance.CheckOptions) ([]compliance.CheckResult, error) {
	var allResults []compliance.CheckResult

	// K8s checks
	k8sResults, _ := runK8sChecks(ctx, opts)
	allResults = append(allResults, k8sResults...)

	// Docker checks
	dockerResults, _ := runDockerChecks(ctx, opts)
	allResults = append(allResults, dockerResults...)

	// File checks
	fileResults, _ := runFileChecks(ctx, opts)
	allResults = append(allResults, fileResults...)

	return allResults, nil
}

func displayResults(results []compliance.CheckResult) {
	if len(results) == 0 {
		output.Success("No issues found!")
		return
	}

	// Group by category
	byCategory := make(map[string][]compliance.CheckResult)
	for _, r := range results {
		byCategory[r.Category] = append(byCategory[r.Category], r)
	}

	// Summary counts
	var passed, failed, warnings, skipped int
	for _, r := range results {
		switch r.Status {
		case compliance.StatusPassed:
			passed++
		case compliance.StatusFailed:
			if r.Severity == "high" || r.Severity == "critical" {
				failed++
			} else {
				warnings++
			}
		case compliance.StatusSkipped:
			skipped++
		}
	}

	// Display by category
	for category, categoryResults := range byCategory {
		output.Newline()
		output.Print(output.Section(category))

		table := output.NewTable(output.TableConfig{
			Headers:    []string{"Status", "Severity", "Rule", "Resource", "Message"},
			ShowBorder: true,
		})

		for _, r := range categoryResults {
			statusIcon := getCheckStatusIcon(r.Status, r.Severity)
			severityBadge := getSeverityBadge(r.Severity)

			table.AddColoredRow(
				[]string{
					statusIcon,
					severityBadge,
					r.RuleID,
					truncateString(r.Resource, 30),
					truncateString(r.Message, 40),
				},
				getCheckRowColors(r),
			)
		}

		table.Render()
	}

	// Summary
	output.Newline()
	output.Print(output.Divider(60))
	output.Newline()
	output.Print(output.Section("Summary"))

	total := passed + failed + warnings + skipped
	output.Printf("  Total Checks: %d\n", total)
	output.Printf("  %s Passed: %d\n", output.SuccessStyle.Render(output.IconSuccess), passed)
	if failed > 0 {
		output.Printf("  %s Failed: %d\n", output.ErrorStyle.Render(output.IconError), failed)
	}
	if warnings > 0 {
		output.Printf("  %s Warnings: %d\n", output.WarningStyle.Render(output.IconWarning), warnings)
	}
	if skipped > 0 {
		output.Printf("  %s Skipped: %d\n", output.MutedStyle.Render(output.IconCross), skipped)
	}

	// Score
	if total > 0 {
		score := float64(passed) / float64(total-skipped) * 100
		bar := output.ProgressBar(int(score), 100, 30)
		output.Printf("\n  Compliance Score: %s %.1f%%\n", bar, score)
	}

	output.Newline()
}

func getCheckStatusIcon(status compliance.CheckStatus, severity string) string {
	switch status {
	case compliance.StatusPassed:
		return output.SuccessStyle.Render(output.IconSuccess)
	case compliance.StatusFailed:
		if severity == "critical" || severity == "high" {
			return output.ErrorStyle.Render(output.IconError)
		}
		return output.WarningStyle.Render(output.IconWarning)
	case compliance.StatusSkipped:
		return output.MutedStyle.Render(output.IconCross)
	default:
		return output.InfoStyle.Render(output.IconInfo)
	}
}

func getSeverityBadge(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return output.Badge("CRIT", "error")
	case "high":
		return output.Badge("HIGH", "error")
	case "medium":
		return output.Badge("MED", "warning")
	case "low":
		return output.Badge("LOW", "info")
	default:
		return severity
	}
}

func getCheckRowColors(r compliance.CheckResult) []tablewriter.Colors {
	var statusColor, severityColor int

	switch r.Status {
	case compliance.StatusPassed:
		statusColor = tablewriter.FgGreenColor
	case compliance.StatusFailed:
		statusColor = tablewriter.FgRedColor
	default:
		statusColor = tablewriter.FgHiBlackColor
	}

	switch strings.ToLower(r.Severity) {
	case "critical":
		severityColor = tablewriter.FgRedColor
	case "high":
		severityColor = tablewriter.FgRedColor
	case "medium":
		severityColor = tablewriter.FgYellowColor
	default:
		severityColor = tablewriter.FgCyanColor
	}

	return []tablewriter.Colors{
		{statusColor},                    // Status
		{severityColor},                  // Severity
		{tablewriter.FgCyanColor},        // Rule
		{tablewriter.FgWhiteColor},       // Resource
		{tablewriter.FgHiBlackColor},     // Message
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}


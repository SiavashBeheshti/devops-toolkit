package completion

import (
	"strings"

	"github.com/spf13/cobra"
)

// ComplianceTargetCompletion provides completion for compliance check/report targets
func ComplianceTargetCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		// Only complete the first argument
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	targets := []string{
		"k8s\tCheck Kubernetes resources",
		"docker\tCheck Docker images and containers",
		"files\tCheck configuration files",
		"all\tRun all available checks",
	}

	var completions []string
	for _, target := range targets {
		parts := strings.Split(target, "\t")
		if strings.HasPrefix(parts[0], toComplete) {
			completions = append(completions, target)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// OutputFormatCompletion provides completion for output format flags
func OutputFormatCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	formats := []string{
		"table\tConsole table output",
		"json\tJSON format",
		"yaml\tYAML format",
	}

	var completions []string
	for _, format := range formats {
		parts := strings.Split(format, "\t")
		if strings.HasPrefix(parts[0], toComplete) {
			completions = append(completions, format)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// ReportFormatCompletion provides completion for report format flags
func ReportFormatCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	formats := []string{
		"table\tConsole table output",
		"json\tJSON format for programmatic use",
		"junit\tJUnit XML format for CI integration",
		"html\tHTML report for sharing",
	}

	var completions []string
	for _, format := range formats {
		parts := strings.Split(format, "\t")
		if strings.HasPrefix(parts[0], toComplete) {
			completions = append(completions, format)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// SeverityCompletion provides completion for severity flags
func SeverityCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	severities := []string{
		"low\tMinimum severity level",
		"medium\tMedium severity level",
		"high\tHigh severity level",
		"critical\tCritical severity level",
	}

	var completions []string
	for _, severity := range severities {
		parts := strings.Split(severity, "\t")
		if strings.HasPrefix(parts[0], toComplete) {
			completions = append(completions, severity)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// PipelineStatusCompletion provides completion for GitLab pipeline status
func PipelineStatusCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	statuses := []string{
		"running\tCurrently running",
		"pending\tWaiting to run",
		"success\tCompleted successfully",
		"failed\tFailed execution",
		"canceled\tCanceled by user",
		"skipped\tSkipped execution",
	}

	var completions []string
	for _, status := range statuses {
		parts := strings.Split(status, "\t")
		if strings.HasPrefix(parts[0], toComplete) {
			completions = append(completions, status)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// SortOptionCompletion provides completion for pod sort options
func PodSortCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	options := []string{
		"name\tSort by pod name",
		"status\tSort by status",
		"age\tSort by creation time",
		"restarts\tSort by restart count",
		"namespace\tSort by namespace",
	}

	var completions []string
	for _, opt := range options {
		parts := strings.Split(opt, "\t")
		if strings.HasPrefix(parts[0], toComplete) {
			completions = append(completions, opt)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// ImageSortCompletion provides completion for image sort options
func ImageSortCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	options := []string{
		"name\tSort by repository name",
		"size\tSort by image size",
		"created\tSort by creation time",
	}

	var completions []string
	for _, opt := range options {
		parts := strings.Split(opt, "\t")
		if strings.HasPrefix(parts[0], toComplete) {
			completions = append(completions, opt)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// NoFileCompletion returns an empty completion that prevents file completion
func NoFileCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}


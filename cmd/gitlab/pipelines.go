package gitlab

import (
	"fmt"
	"strings"

	"github.com/beheshti/devops-toolkit/pkg/gitlabclient"
	"github.com/beheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newPipelinesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pipelines",
		Aliases: []string{"pl", "pipe"},
		Short:   "List and manage pipelines",
		Long: `List and manage GitLab CI/CD pipelines.

Features:
  • Color-coded pipeline status
  • Duration and timing information
  • Branch and commit details
  • Filtering by status and ref`,
		RunE: runPipelines,
	}

	cmd.Flags().StringP("status", "s", "", "Filter by status (running, pending, success, failed, canceled)")
	cmd.Flags().StringP("ref", "r", "", "Filter by branch/tag ref")
	cmd.Flags().IntP("limit", "n", 20, "Number of pipelines to show")
	cmd.Flags().Bool("all", false, "Show pipelines from all branches")

	return cmd
}

func runPipelines(cmd *cobra.Command, args []string) error {
	output.StartSpinner("Fetching pipelines...")

	client, projectID, err := getClient(cmd)
	if err != nil {
		output.SpinnerError("Failed to connect to GitLab")
		return err
	}

	status, _ := cmd.Flags().GetString("status")
	ref, _ := cmd.Flags().GetString("ref")
	limit, _ := cmd.Flags().GetInt("limit")

	pipelines, err := client.ListPipelines(projectID, gitlabclient.PipelineFilter{
		Status: status,
		Ref:    ref,
		Limit:  limit,
	})
	if err != nil {
		output.SpinnerError("Failed to fetch pipelines")
		return fmt.Errorf("failed to list pipelines: %w", err)
	}

	output.SpinnerSuccess(fmt.Sprintf("Found %d pipelines", len(pipelines)))
	output.Newline()

	if len(pipelines) == 0 {
		output.Info("No pipelines found matching the criteria")
		return nil
	}

	// Build table
	table := output.NewTable(output.TableConfig{
		Title:      "CI/CD Pipelines",
		Headers:    []string{"ID", "Status", "Ref", "Commit", "Created", "Duration"},
		ShowBorder: true,
	})

	// Status counts
	statusCounts := make(map[string]int)

	for _, pl := range pipelines {
		statusCounts[pl.Status]++

		statusIcon := getPipelineStatusIcon(pl.Status)
		status := fmt.Sprintf("%s %s", statusIcon, pl.Status)

		commit := pl.SHA
		if len(commit) > 8 {
			commit = commit[:8]
		}

		ref := pl.Ref
		if len(ref) > 20 {
			ref = ref[:17] + "..."
		}

		table.AddColoredRow(
			[]string{
				fmt.Sprintf("#%d", pl.ID),
				status,
				ref,
				commit,
				formatDuration(pl.CreatedAt),
				pl.Duration,
			},
			getPipelineRowColors(pl.Status),
		)
	}

	table.Render()

	// Summary
	output.Newline()
	output.Print(output.Section("Pipeline Summary"))
	for status, count := range statusCounts {
		icon := getPipelineStatusIcon(status)
		output.Printf("  %s %s: %d\n", icon, status, count)
	}
	output.Newline()

	return nil
}

func getPipelineStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "success", "passed":
		return output.SuccessStyle.Render(output.IconSuccess)
	case "failed":
		return output.ErrorStyle.Render(output.IconError)
	case "running":
		return output.InfoStyle.Render(output.IconRunning)
	case "pending", "waiting_for_resource", "preparing":
		return output.WarningStyle.Render(output.IconPending)
	case "canceled", "cancelled", "skipped":
		return output.MutedStyle.Render(output.IconCross)
	default:
		return output.InfoStyle.Render(output.IconInfo)
	}
}

func getPipelineRowColors(status string) []tablewriter.Colors {
	var statusColor int
	switch strings.ToLower(status) {
	case "success", "passed":
		statusColor = tablewriter.FgGreenColor
	case "failed":
		statusColor = tablewriter.FgRedColor
	case "running":
		statusColor = tablewriter.FgBlueColor
	case "pending", "waiting_for_resource":
		statusColor = tablewriter.FgYellowColor
	case "canceled", "cancelled", "skipped":
		statusColor = tablewriter.FgHiBlackColor
	default:
		statusColor = tablewriter.FgWhiteColor
	}

	return []tablewriter.Colors{
		{tablewriter.FgCyanColor},        // ID
		{tablewriter.Bold, statusColor},  // Status
		{tablewriter.FgMagentaColor},     // Ref
		{tablewriter.FgHiBlackColor},     // Commit
		{tablewriter.FgHiBlackColor},     // Created
		{tablewriter.FgWhiteColor},       // Duration
	}
}

func formatDuration(timeStr string) string {
	// Parse and format time
	if timeStr == "" {
		return "-"
	}
	// Simplified - just return the time string
	return timeStr
}


package gitlab

import (
	"fmt"
	"strings"

	"github.com/SiavashBeheshti/devops-toolkit/pkg/gitlabclient"
	"github.com/SiavashBeheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newJobsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jobs",
		Aliases: []string{"j"},
		Short:   "List and manage pipeline jobs",
		Long: `List and manage GitLab CI/CD pipeline jobs.

Features:
  • Color-coded job status
  • Stage grouping
  • Duration tracking
  • Log access`,
		RunE: runJobs,
	}

	cmd.Flags().IntP("pipeline", "i", 0, "Pipeline ID (required)")
	cmd.Flags().StringP("status", "s", "", "Filter by status")
	cmd.Flags().String("stage", "", "Filter by stage")
	cmd.Flags().Bool("failed", false, "Show only failed jobs")

	return cmd
}

func runJobs(cmd *cobra.Command, args []string) error {
	pipelineID, _ := cmd.Flags().GetInt("pipeline")
	if pipelineID == 0 {
		return fmt.Errorf("pipeline ID is required (use -i flag)")
	}

	output.StartSpinner("Fetching jobs...")

	client, projectID, err := getClient(cmd)
	if err != nil {
		output.SpinnerError("Failed to connect to GitLab")
		return err
	}

	status, _ := cmd.Flags().GetString("status")
	stage, _ := cmd.Flags().GetString("stage")
	failedOnly, _ := cmd.Flags().GetBool("failed")

	if failedOnly {
		status = "failed"
	}

	jobs, err := client.ListPipelineJobs(projectID, pipelineID, gitlabclient.JobFilter{
		Status: status,
		Stage:  stage,
	})
	if err != nil {
		output.SpinnerError("Failed to fetch jobs")
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	output.SpinnerSuccess(fmt.Sprintf("Found %d jobs", len(jobs)))
	output.Newline()

	if len(jobs) == 0 {
		output.Info("No jobs found matching the criteria")
		return nil
	}

	// Group jobs by stage
	stageJobs := make(map[string][]gitlabclient.JobInfo)
	stageOrder := []string{}
	for _, job := range jobs {
		if _, exists := stageJobs[job.Stage]; !exists {
			stageOrder = append(stageOrder, job.Stage)
		}
		stageJobs[job.Stage] = append(stageJobs[job.Stage], job)
	}

	output.Header(fmt.Sprintf("Pipeline #%d Jobs", pipelineID))

	// Display jobs grouped by stage
	for _, stageName := range stageOrder {
		stageJobsList := stageJobs[stageName]

		// Count stage status
		var passed, failed, running, pending int
		for _, j := range stageJobsList {
			switch strings.ToLower(j.Status) {
			case "success", "passed":
				passed++
			case "failed":
				failed++
			case "running":
				running++
			case "pending":
				pending++
			}
		}

		// Stage header with summary
		stageIcon := getStageIcon(passed, failed, running, len(stageJobsList))
		output.Printf("\n  %s %s\n", stageIcon, output.InfoStyle.Render(stageName))
		output.Printf("     %s\n", output.MutedStyle.Render(strings.Repeat("─", 50)))

		for _, job := range stageJobsList {
			statusIcon := getJobStatusIcon(job.Status)
			duration := job.Duration
			if duration == "" {
				duration = "-"
			}

			output.Printf("     %s %-30s %s  %s\n",
				statusIcon,
				job.Name,
				output.MutedStyle.Render(duration),
				getJobStatusBadge(job.Status))
		}
	}

	// Summary table
	output.Newline()
	summaryTable := output.NewTable(output.TableConfig{
		Title:      "Job Summary",
		Headers:    []string{"Status", "Count", ""},
		ShowBorder: true,
	})

	statusCounts := make(map[string]int)
	for _, job := range jobs {
		statusCounts[job.Status]++
	}

	for status, count := range statusCounts {
		icon := getJobStatusIcon(status)
		summaryTable.AddColoredRow(
			[]string{status, fmt.Sprintf("%d", count), icon},
			getJobSummaryColors(status),
		)
	}

	output.Newline()
	summaryTable.Render()
	output.Newline()

	return nil
}

func getStageIcon(passed, failed, running, total int) string {
	if failed > 0 {
		return output.ErrorStyle.Render("●")
	}
	if running > 0 {
		return output.InfoStyle.Render("●")
	}
	if passed == total {
		return output.SuccessStyle.Render("●")
	}
	return output.WarningStyle.Render("●")
}

func getJobStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "success", "passed":
		return output.SuccessStyle.Render(output.IconSuccess)
	case "failed":
		return output.ErrorStyle.Render(output.IconError)
	case "running":
		return output.InfoStyle.Render(output.IconRunning)
	case "pending", "created":
		return output.WarningStyle.Render(output.IconPending)
	case "canceled", "cancelled", "skipped":
		return output.MutedStyle.Render(output.IconCross)
	case "manual":
		return output.InfoStyle.Render("▶")
	default:
		return output.MutedStyle.Render(output.IconBullet)
	}
}

func getJobStatusBadge(status string) string {
	switch strings.ToLower(status) {
	case "success", "passed":
		return output.Badge("PASSED", "success")
	case "failed":
		return output.Badge("FAILED", "error")
	case "running":
		return output.Badge("RUNNING", "info")
	case "pending":
		return output.Badge("PENDING", "warning")
	case "manual":
		return output.Badge("MANUAL", "info")
	default:
		return ""
	}
}

func getJobSummaryColors(status string) []tablewriter.Colors {
	var color int
	switch strings.ToLower(status) {
	case "success", "passed":
		color = tablewriter.FgGreenColor
	case "failed":
		color = tablewriter.FgRedColor
	case "running":
		color = tablewriter.FgBlueColor
	case "pending":
		color = tablewriter.FgYellowColor
	default:
		color = tablewriter.FgHiBlackColor
	}

	return []tablewriter.Colors{
		{color},
		{tablewriter.FgWhiteColor},
		{color},
	}
}

package gitlab

import (
	"fmt"
	"os"

	"github.com/SiavashBeheshti/devops-toolkit/pkg/gitlabclient"
	"github.com/SiavashBeheshti/devops-toolkit/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show project CI/CD status overview",
		Long: `Display an overview of CI/CD status for a project.

Shows:
  • Latest pipeline status per branch
  • Recent pipeline history
  • Job success rates
  • Environment deployments`,
		RunE: runStatus,
	}

	cmd.Flags().Bool("all-branches", false, "Show status for all branches")

	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	output.StartSpinner("Fetching project status...")

	client, projectID, err := getClient(cmd)
	if err != nil {
		output.SpinnerError("Failed to connect to GitLab")
		return err
	}

	// Get project info
	project, err := client.GetProject(projectID)
	if err != nil {
		output.SpinnerError("Failed to fetch project")
		return fmt.Errorf("failed to get project: %w", err)
	}

	output.SpinnerSuccess("Project found")
	output.Newline()

	output.Header(fmt.Sprintf("Project: %s", project.Name))
	output.Printf("  %s\n", output.KeyValue("Path", project.PathWithNamespace))
	output.Printf("  %s\n", output.KeyValue("Default Branch", project.DefaultBranch))
	output.Printf("  %s\n", output.KeyValue("Web URL", project.WebURL))

	// Get latest pipeline on default branch
	output.Newline()
	output.Print(output.Section("Latest Pipeline"))

	latestPipeline, err := client.GetLatestPipeline(projectID, project.DefaultBranch)
	if err != nil {
		output.Warning("No pipelines found")
	} else {
		statusIcon := getPipelineStatusIcon(latestPipeline.Status)
		output.Printf("  %s Pipeline #%d: %s\n", statusIcon, latestPipeline.ID, latestPipeline.Status)
		output.Printf("     Ref: %s\n", output.InfoStyle.Render(latestPipeline.Ref))
		output.Printf("     Commit: %s\n", output.MutedStyle.Render(latestPipeline.SHA[:8]))
		output.Printf("     Duration: %s\n", latestPipeline.Duration)

		// Get jobs for this pipeline
		jobs, _ := client.ListPipelineJobs(projectID, latestPipeline.ID, gitlabclient.JobFilter{})
		if len(jobs) > 0 {
			output.Newline()
			output.Print(output.SubSection("Jobs"))

			for _, job := range jobs {
				icon := getJobStatusIcon(job.Status)
				output.Printf("     %s %s (%s)\n", icon, job.Name, job.Stage)
			}
		}
	}

	// Pipeline statistics
	output.Newline()
	output.Print(output.Section("Pipeline Statistics (Last 30 Days)"))

	stats, err := client.GetPipelineStats(projectID)
	if err == nil {
		total := stats.Success + stats.Failed + stats.Other
		successRate := float64(0)
		if total > 0 {
			successRate = float64(stats.Success) / float64(total) * 100
		}

		output.Printf("  Total Pipelines: %d\n", total)
		output.Printf("  %s Success: %d (%.1f%%)\n",
			output.SuccessStyle.Render(output.IconSuccess),
			stats.Success, successRate)
		output.Printf("  %s Failed: %d\n",
			output.ErrorStyle.Render(output.IconError),
			stats.Failed)
		output.Printf("  Average Duration: %s\n", stats.AvgDuration)

		// Visual bar
		if total > 0 {
			bar := output.ProgressBar(int(successRate), 100, 30)
			output.Printf("\n  Success Rate: %s\n", bar)
		}
	}

	// Environments
	output.Newline()
	output.Print(output.Section("Environments"))

	environments, err := client.ListEnvironments(projectID)
	if err == nil && len(environments) > 0 {
		for _, env := range environments {
			icon := output.SuccessStyle.Render(output.IconSuccess)
			if env.State != "available" {
				icon = output.MutedStyle.Render(output.IconPending)
			}
			output.Printf("  %s %s\n", icon, env.Name)
			if env.LastDeployment != "" {
				output.Printf("     Last deployment: %s\n",
					output.MutedStyle.Render(env.LastDeployment))
			}
			if env.ExternalURL != "" {
				output.Printf("     URL: %s\n",
					output.InfoStyle.Render(env.ExternalURL))
			}
		}
	} else {
		output.Muted("  No environments configured")
	}

	output.Newline()
	return nil
}

func getClient(cmd *cobra.Command) (*gitlabclient.Client, string, error) {
	token := cmd.Flag("token").Value.String()
	if token == "" {
		token = os.Getenv("GITLAB_TOKEN")
	}
	if token == "" {
		token = viper.GetString("gitlab.token")
	}
	if token == "" {
		return nil, "", fmt.Errorf("GitLab token required (use --token flag or GITLAB_TOKEN env)")
	}

	url := cmd.Flag("url").Value.String()
	if url == "" {
		url = os.Getenv("GITLAB_URL")
	}
	if url == "" {
		url = viper.GetString("gitlab.url")
	}
	if url == "" {
		url = "https://gitlab.com"
	}

	projectID := cmd.Flag("project").Value.String()
	if projectID == "" {
		projectID = os.Getenv("GITLAB_PROJECT")
	}
	if projectID == "" {
		projectID = viper.GetString("gitlab.project")
	}
	if projectID == "" {
		// Try to detect from git remote
		projectID = detectProjectFromGit()
	}
	if projectID == "" {
		return nil, "", fmt.Errorf("project ID required (use --project flag or GITLAB_PROJECT env)")
	}

	client, err := gitlabclient.NewClient(url, token)
	if err != nil {
		return nil, "", err
	}

	return client, projectID, nil
}

func detectProjectFromGit() string {
	// Try to detect project from git remote
	// This is a simplified implementation
	return ""
}

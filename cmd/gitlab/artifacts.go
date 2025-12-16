package gitlab

import (
	"fmt"

	"github.com/beheshti/devops-toolkit/pkg/gitlabclient"
	"github.com/beheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newArtifactsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "artifacts",
		Aliases: []string{"art"},
		Short:   "Manage pipeline artifacts",
		Long: `View and download pipeline artifacts.

Features:
  • List artifacts by job
  • Download artifacts
  • Size information
  • Expiration tracking`,
		RunE: runArtifacts,
	}

	cmd.Flags().IntP("pipeline", "i", 0, "Pipeline ID")
	cmd.Flags().IntP("job", "j", 0, "Job ID")
	cmd.Flags().String("download", "", "Download artifact to path")

	return cmd
}

func runArtifacts(cmd *cobra.Command, args []string) error {
	pipelineID, _ := cmd.Flags().GetInt("pipeline")
	jobID, _ := cmd.Flags().GetInt("job")

	output.StartSpinner("Fetching artifacts...")

	client, projectID, err := getClient(cmd)
	if err != nil {
		output.SpinnerError("Failed to connect to GitLab")
		return err
	}

	var artifacts []gitlabclient.ArtifactInfo

	if jobID > 0 {
		// Get artifacts for specific job
		artifact, err := client.GetJobArtifacts(projectID, jobID)
		if err != nil {
			output.SpinnerError("Failed to fetch artifacts")
			return fmt.Errorf("failed to get artifacts: %w", err)
		}
		if artifact != nil {
			artifacts = append(artifacts, *artifact)
		}
	} else if pipelineID > 0 {
		// Get all artifacts from pipeline
		artifacts, err = client.ListPipelineArtifacts(projectID, pipelineID)
		if err != nil {
			output.SpinnerError("Failed to fetch artifacts")
			return fmt.Errorf("failed to list artifacts: %w", err)
		}
	} else {
		output.SpinnerError("Missing required flags")
		return fmt.Errorf("either --pipeline or --job is required")
	}

	output.SpinnerSuccess(fmt.Sprintf("Found %d artifacts", len(artifacts)))
	output.Newline()

	if len(artifacts) == 0 {
		output.Info("No artifacts found")
		return nil
	}

	// Build table
	table := output.NewTable(output.TableConfig{
		Title:      "Pipeline Artifacts",
		Headers:    []string{"Job", "Name", "Size", "Expires"},
		ShowBorder: true,
	})

	var totalSize int64
	for _, art := range artifacts {
		totalSize += art.Size

		table.AddColoredRow(
			[]string{
				art.JobName,
				art.Filename,
				formatArtifactSize(art.Size),
				art.ExpireAt,
			},
			[]tablewriter.Colors{
				{tablewriter.FgCyanColor},
				{tablewriter.FgWhiteColor},
				{tablewriter.FgYellowColor},
				{tablewriter.FgHiBlackColor},
			},
		)
	}

	table.Render()

	// Summary
	output.Newline()
	output.Printf("Total artifact size: %s\n", formatArtifactSize(totalSize))
	output.Newline()

	return nil
}

func formatArtifactSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}


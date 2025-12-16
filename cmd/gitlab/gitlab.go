package gitlab

import (
	"github.com/spf13/cobra"
)

// NewGitLabCmd creates the gitlab command
func NewGitLabCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gitlab",
		Aliases: []string{"gl"},
		Short:   "GitLab CI/CD operations",
		Long: `GitLab CI/CD pipeline management and monitoring.

Provides enhanced visibility into your GitLab pipelines
with beautiful output and powerful filtering options.`,
	}

	// Add subcommands
	cmd.AddCommand(newPipelinesCmd())
	cmd.AddCommand(newJobsCmd())
	cmd.AddCommand(newTriggerCmd())
	cmd.AddCommand(newArtifactsCmd())
	cmd.AddCommand(newStatusCmd())

	// Persistent flags
	cmd.PersistentFlags().String("token", "", "GitLab access token (or set GITLAB_TOKEN)")
	cmd.PersistentFlags().String("url", "https://gitlab.com", "GitLab instance URL")
	cmd.PersistentFlags().StringP("project", "p", "", "Project ID or path")

	return cmd
}


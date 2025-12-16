package docker

import (
	"github.com/spf13/cobra"
)

// NewDockerCmd creates the docker command
func NewDockerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "docker",
		Aliases: []string{"d", "container"},
		Short:   "Docker container operations",
		Long: `Docker container and image operations with enhanced visibility.

Provides beautiful, informative output for Docker operations
including container stats, image analysis, and cleanup tools.`,
	}

	// Add subcommands
	cmd.AddCommand(newContainersCmd())
	cmd.AddCommand(newImagesCmd())
	cmd.AddCommand(newStatsCmd())
	cmd.AddCommand(newCleanCmd())
	cmd.AddCommand(newInspectCmd())
	cmd.AddCommand(newLogsCmd())

	// Persistent flags
	cmd.PersistentFlags().StringP("host", "H", "", "Docker host to connect to")

	return cmd
}


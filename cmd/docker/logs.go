package docker

import (
	"context"
	"fmt"

	"github.com/beheshti/devops-toolkit/pkg/completion"
	"github.com/beheshti/devops-toolkit/pkg/docker"
	"github.com/beheshti/devops-toolkit/pkg/output"
	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs [container]",
		Short: "View container logs with highlighting",
		Long: `View container logs with syntax highlighting and filtering.

Features:
  • Error/warning highlighting
  • JSON log parsing
  • Timestamp formatting
  • Log level filtering`,
		Args:              cobra.ExactArgs(1),
		RunE:              runLogs,
		ValidArgsFunction: completion.RunningContainerCompletion,
	}

	cmd.Flags().IntP("tail", "n", 100, "Number of lines to show")
	cmd.Flags().BoolP("follow", "f", false, "Follow log output")
	cmd.Flags().Bool("timestamps", false, "Show timestamps")
	cmd.Flags().String("since", "", "Show logs since timestamp (e.g. 2023-01-01T00:00:00)")
	cmd.Flags().String("until", "", "Show logs until timestamp")
	cmd.Flags().String("level", "", "Filter by log level (error, warn, info, debug)")

	// Register flag completions
	_ = cmd.RegisterFlagCompletionFunc("level", completion.LogLevelCompletion)

	return cmd
}

func runLogs(cmd *cobra.Command, args []string) error {
	containerID := args[0]

	client, err := docker.NewClient()
	if err != nil {
		output.Error("Failed to connect to Docker")
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	tail, _ := cmd.Flags().GetInt("tail")
	follow, _ := cmd.Flags().GetBool("follow")
	timestamps, _ := cmd.Flags().GetBool("timestamps")
	since, _ := cmd.Flags().GetString("since")
	until, _ := cmd.Flags().GetString("until")
	level, _ := cmd.Flags().GetString("level")

	opts := docker.LogOptions{
		Tail:       tail,
		Follow:     follow,
		Timestamps: timestamps,
		Since:      since,
		Until:      until,
		Level:      level,
	}

	output.Header(fmt.Sprintf("Logs: %s", containerID))

	if follow {
		output.Info("Following logs... (Ctrl+C to stop)")
	}

	err = client.StreamLogs(ctx, containerID, opts, func(line docker.LogLine) {
		printLogLine(line)
	})

	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}

	return nil
}

func printLogLine(line docker.LogLine) {
	var prefix string

	// Timestamp
	if line.Timestamp != "" {
		prefix = output.MutedStyle.Render(line.Timestamp) + " "
	}

	// Stream indicator
	if line.Stream == "stderr" {
		prefix += output.ErrorStyle.Render("ERR") + " "
	}

	// Color based on detected level
	var content string
	switch line.Level {
	case "error", "fatal", "panic":
		content = output.ErrorStyle.Render(line.Content)
	case "warn", "warning":
		content = output.WarningStyle.Render(line.Content)
	case "info":
		content = output.InfoStyle.Render(line.Content)
	case "debug", "trace":
		content = output.MutedStyle.Render(line.Content)
	default:
		content = line.Content
	}

	fmt.Printf("%s%s\n", prefix, content)
}


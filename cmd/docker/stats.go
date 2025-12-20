package docker

import (
	"context"
	"fmt"

	"github.com/SiavashBeheshti/devops-toolkit/pkg/docker"
	"github.com/SiavashBeheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show container statistics",
		Long: `Display real-time container resource usage statistics.

Features:
  • CPU and Memory usage with progress bars
  • Network I/O statistics
  • Block I/O statistics
  • PIDs count`,
		RunE: runStats,
	}

	cmd.Flags().Bool("no-stream", true, "Disable streaming stats (show once)")
	cmd.Flags().StringP("format", "f", "table", "Output format (table, json)")

	return cmd
}

func runStats(cmd *cobra.Command, args []string) error {
	output.StartSpinner("Fetching container stats...")

	client, err := docker.NewClient()
	if err != nil {
		output.SpinnerError("Failed to connect to Docker")
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Get running containers
	containers, err := client.ListContainers(ctx, false)
	if err != nil {
		output.SpinnerError("Failed to list containers")
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		output.SpinnerError("No running containers")
		output.Info("No running containers to show stats for")
		return nil
	}

	// Get stats for each container
	stats, err := client.GetContainerStats(ctx, containers)
	if err != nil {
		output.SpinnerError("Failed to get stats")
		return fmt.Errorf("failed to get container stats: %w", err)
	}

	output.SpinnerSuccess(fmt.Sprintf("Stats for %d containers", len(stats)))
	output.Newline()

	// Build table
	table := output.NewTable(output.TableConfig{
		Title:      "Container Statistics",
		Headers:    []string{"Container", "CPU %", "Memory", "Mem %", "Net I/O", "Block I/O", "PIDs"},
		ShowBorder: true,
	})

	var totalCPU, totalMemPercent float64
	var totalMemUsage, totalMemLimit int64

	for _, stat := range stats {
		cpuPercent := fmt.Sprintf("%.1f%%", stat.CPUPercent)
		memPercent := fmt.Sprintf("%.1f%%", stat.MemoryPercent)
		memUsage := fmt.Sprintf("%s / %s", formatSize(stat.MemoryUsage), formatSize(stat.MemoryLimit))
		netIO := fmt.Sprintf("%s / %s", formatSize(stat.NetInput), formatSize(stat.NetOutput))
		blockIO := fmt.Sprintf("%s / %s", formatSize(stat.BlockInput), formatSize(stat.BlockOutput))

		table.AddColoredRow(
			[]string{
				truncateName(stat.Name, 20),
				cpuPercent,
				memUsage,
				memPercent,
				netIO,
				blockIO,
				fmt.Sprintf("%d", stat.PIDs),
			},
			getStatsRowColors(stat),
		)

		totalCPU += stat.CPUPercent
		totalMemPercent += stat.MemoryPercent
		totalMemUsage += stat.MemoryUsage
		totalMemLimit += stat.MemoryLimit
	}

	table.Render()

	// Summary
	output.Newline()
	output.Print(output.Section("Resource Summary"))
	output.Printf("  Total CPU: %.1f%%\n", totalCPU)
	output.Printf("  Total Memory: %s / %s (%.1f%%)\n",
		formatSize(totalMemUsage),
		formatSize(totalMemLimit),
		totalMemPercent/float64(len(stats)))

	// Alerts for high usage
	output.Newline()
	hasAlerts := false
	for _, stat := range stats {
		if stat.CPUPercent > 80 {
			if !hasAlerts {
				output.Print(output.Section("Alerts"))
				hasAlerts = true
			}
			output.Printf("  %s %s: High CPU usage (%.1f%%)\n",
				output.WarningStyle.Render(output.IconWarning),
				stat.Name, stat.CPUPercent)
		}
		if stat.MemoryPercent > 80 {
			if !hasAlerts {
				output.Print(output.Section("Alerts"))
				hasAlerts = true
			}
			output.Printf("  %s %s: High memory usage (%.1f%%)\n",
				output.WarningStyle.Render(output.IconWarning),
				stat.Name, stat.MemoryPercent)
		}
	}

	if !hasAlerts {
		output.Success("All containers within normal resource limits")
	}

	output.Newline()
	return nil
}

func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen-3] + "..."
}

func getStatsRowColors(stat docker.ContainerStats) []tablewriter.Colors {
	cpuColor := getResourceColorByPercent(stat.CPUPercent)
	memColor := getResourceColorByPercent(stat.MemoryPercent)

	return []tablewriter.Colors{
		{tablewriter.FgCyanColor},    // Name
		{cpuColor},                   // CPU
		{tablewriter.FgWhiteColor},   // Mem Usage
		{memColor},                   // Mem %
		{tablewriter.FgHiBlackColor}, // Net I/O
		{tablewriter.FgHiBlackColor}, // Block I/O
		{tablewriter.FgWhiteColor},   // PIDs
	}
}

func getResourceColorByPercent(percent float64) int {
	switch {
	case percent > 90:
		return tablewriter.FgRedColor
	case percent > 70:
		return tablewriter.FgYellowColor
	default:
		return tablewriter.FgGreenColor
	}
}

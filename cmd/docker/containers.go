package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/SiavashBeheshti/devops-toolkit/pkg/docker"
	"github.com/SiavashBeheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newContainersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "containers",
		Aliases: []string{"ps", "list"},
		Short:   "List containers with enhanced info",
		Long: `List Docker containers with enhanced information and formatting.

Features:
  • Color-coded status indicators
  • Resource usage display
  • Port mapping visualization
  • Health check status`,
		RunE: runContainers,
	}

	cmd.Flags().BoolP("all", "a", false, "Show all containers (including stopped)")
	cmd.Flags().Bool("wide", false, "Show additional information")
	cmd.Flags().StringP("filter", "f", "", "Filter containers (name, status, label)")
	cmd.Flags().Bool("size", false, "Show container sizes")

	return cmd
}

func runContainers(cmd *cobra.Command, args []string) error {
	output.StartSpinner("Fetching containers...")

	client, err := docker.NewClient()
	if err != nil {
		output.SpinnerError("Failed to connect to Docker")
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	showAll, _ := cmd.Flags().GetBool("all")
	wide, _ := cmd.Flags().GetBool("wide")
	showSize, _ := cmd.Flags().GetBool("size")

	containers, err := client.ListContainers(ctx, showAll)
	if err != nil {
		output.SpinnerError("Failed to list containers")
		return fmt.Errorf("failed to list containers: %w", err)
	}

	output.SpinnerSuccess(fmt.Sprintf("Found %d containers", len(containers)))
	output.Newline()

	if len(containers) == 0 {
		output.Info("No containers found")
		return nil
	}

	// Build table
	headers := []string{"Container ID", "Image", "Status", "Ports", "Name"}
	if wide {
		headers = append(headers, "Command", "Created")
	}
	if showSize {
		headers = append(headers, "Size")
	}

	table := output.NewTable(output.TableConfig{
		Title:      "Docker Containers",
		Headers:    headers,
		ShowBorder: true,
	})

	// Count by status
	statusCounts := make(map[string]int)

	for _, container := range containers {
		statusCounts[container.State]++

		// Format ports
		ports := formatPorts(container.Ports)

		// Format status with health
		status := container.Status
		if container.Health != "" {
			status = fmt.Sprintf("%s (%s)", container.Status, container.Health)
		}

		row := []string{
			truncateID(container.ID),
			truncateImage(container.Image),
			status,
			ports,
			strings.TrimPrefix(container.Name, "/"),
		}

		if wide {
			row = append(row, truncate(container.Command, 30), container.Created)
		}
		if showSize {
			row = append(row, container.Size)
		}

		colors := getContainerRowColors(container, wide, showSize)
		table.AddColoredRow(row, colors)
	}

	table.Render()

	// Summary
	output.Newline()
	output.Print(output.Section("Container Summary"))
	for state, count := range statusCounts {
		icon := output.StatusIcon(state)
		output.Printf("  %s %s: %d\n", icon, state, count)
	}
	output.Newline()

	return nil
}

func truncateID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func truncateImage(image string) string {
	if len(image) > 35 {
		return image[:32] + "..."
	}
	return image
}

func formatPorts(ports []docker.PortMapping) string {
	if len(ports) == 0 {
		return "-"
	}

	var mappings []string
	for _, p := range ports {
		if p.PublicPort > 0 {
			mappings = append(mappings, fmt.Sprintf("%d->%d/%s", p.PublicPort, p.PrivatePort, p.Type))
		} else {
			mappings = append(mappings, fmt.Sprintf("%d/%s", p.PrivatePort, p.Type))
		}
	}

	result := strings.Join(mappings, ", ")
	if len(result) > 30 {
		return result[:27] + "..."
	}
	return result
}

func getContainerRowColors(container docker.ContainerInfo, wide, showSize bool) []tablewriter.Colors {
	var statusColor int
	switch container.State {
	case "running":
		statusColor = tablewriter.FgGreenColor
	case "paused":
		statusColor = tablewriter.FgYellowColor
	case "restarting":
		statusColor = tablewriter.FgYellowColor
	case "exited", "dead":
		statusColor = tablewriter.FgRedColor
	default:
		statusColor = tablewriter.FgWhiteColor
	}

	// Health color
	if container.Health == "unhealthy" {
		statusColor = tablewriter.FgRedColor
	}

	colors := []tablewriter.Colors{
		{tablewriter.FgCyanColor},       // ID
		{tablewriter.FgWhiteColor},      // Image
		{tablewriter.Bold, statusColor}, // Status
		{tablewriter.FgHiBlackColor},    // Ports
		{tablewriter.FgMagentaColor},    // Name
	}

	if wide {
		colors = append(colors,
			tablewriter.Colors{tablewriter.FgHiBlackColor}, // Command
			tablewriter.Colors{tablewriter.FgHiBlackColor}, // Created
		)
	}
	if showSize {
		colors = append(colors, tablewriter.Colors{tablewriter.FgYellowColor})
	}

	return colors
}

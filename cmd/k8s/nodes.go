package k8s

import (
	"context"
	"fmt"

	"github.com/SiavashBeheshti/devops-toolkit/pkg/k8s"
	"github.com/SiavashBeheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newNodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodes",
		Short: "List and analyze cluster nodes",
		Long: `Display cluster nodes with resource utilization and status.

Features:
  • Resource utilization bars
  • Condition status indicators
  • Taints and labels display
  • Role identification`,
		RunE: runNodes,
	}

	cmd.Flags().Bool("wide", false, "Show additional information")
	cmd.Flags().Bool("resources", false, "Show detailed resource info")

	return cmd
}

func runNodes(cmd *cobra.Command, args []string) error {
	output.StartSpinner("Fetching nodes...")

	client, err := k8s.NewClient(
		cmd.Flag("kubeconfig").Value.String(),
		cmd.Flag("context").Value.String(),
	)
	if err != nil {
		output.SpinnerError("Failed to connect to cluster")
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx := context.Background()
	wide, _ := cmd.Flags().GetBool("wide")
	showResources, _ := cmd.Flags().GetBool("resources")

	nodes, err := client.ListNodes(ctx)
	if err != nil {
		output.SpinnerError("Failed to fetch nodes")
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	output.SpinnerSuccess(fmt.Sprintf("Found %d nodes", len(nodes)))
	output.Newline()

	// Build headers
	headers := []string{"Name", "Status", "Roles", "Age", "Version"}
	if showResources {
		headers = append(headers, "CPU", "Memory")
	}
	if wide {
		headers = append(headers, "Internal IP", "OS", "Kernel")
	}

	table := output.NewTable(output.TableConfig{
		Title:      "Cluster Nodes",
		Headers:    headers,
		ShowBorder: true,
	})

	var readyCount, notReadyCount int

	for _, node := range nodes {
		status := "Ready"
		statusIcon := output.IconSuccess
		if !node.Ready {
			status = "NotReady"
			statusIcon = output.IconError
			notReadyCount++
		} else {
			readyCount++
		}

		row := []string{
			node.Name,
			fmt.Sprintf("%s %s", statusIcon, status),
			node.Roles,
			formatAge(node.CreationTime),
			node.KubeletVersion,
		}

		if showResources {
			cpuBar := output.ProgressBar(int(node.CPUUsagePercent), 100, 15)
			memBar := output.ProgressBar(int(node.MemoryUsagePercent), 100, 15)
			row = append(row, cpuBar, memBar)
		}

		if wide {
			row = append(row, node.InternalIP, node.OSImage, node.KernelVersion)
		}

		colors := getNodeRowColors(node, showResources, wide)
		table.AddColoredRow(row, colors)
	}

	table.Render()

	// Summary
	output.Newline()
	output.Print(output.Section("Node Summary"))
	output.Printf("  %s Ready: %d\n", output.SuccessStyle.Render(output.IconSuccess), readyCount)
	if notReadyCount > 0 {
		output.Printf("  %s NotReady: %d\n", output.ErrorStyle.Render(output.IconError), notReadyCount)
	}

	// Show conditions for problematic nodes
	var problemNodes []k8s.NodeInfo
	for _, node := range nodes {
		if !node.Ready || node.MemoryPressure || node.DiskPressure || node.PIDPressure {
			problemNodes = append(problemNodes, node)
		}
	}

	if len(problemNodes) > 0 {
		output.Newline()
		output.Print(output.Section("Node Conditions"))

		for _, node := range problemNodes {
			output.Printf("\n  %s %s:\n", output.WarningStyle.Render(output.IconWarning), node.Name)
			if !node.Ready {
				output.Printf("    %s NotReady\n", output.ErrorStyle.Render(output.IconError))
			}
			if node.MemoryPressure {
				output.Printf("    %s MemoryPressure\n", output.WarningStyle.Render(output.IconWarning))
			}
			if node.DiskPressure {
				output.Printf("    %s DiskPressure\n", output.WarningStyle.Render(output.IconWarning))
			}
			if node.PIDPressure {
				output.Printf("    %s PIDPressure\n", output.WarningStyle.Render(output.IconWarning))
			}
		}
	}

	output.Newline()
	return nil
}

func getNodeRowColors(node k8s.NodeInfo, showResources, wide bool) []tablewriter.Colors {
	var statusColor int
	if node.Ready {
		statusColor = tablewriter.FgGreenColor
	} else {
		statusColor = tablewriter.FgRedColor
	}

	colors := []tablewriter.Colors{
		{tablewriter.FgCyanColor},       // name
		{tablewriter.Bold, statusColor}, // status
		{tablewriter.FgMagentaColor},    // roles
		{tablewriter.FgHiBlackColor},    // age
		{tablewriter.FgWhiteColor},      // version
	}

	if showResources {
		cpuColor := getResourceColor(node.CPUUsagePercent)
		memColor := getResourceColor(node.MemoryUsagePercent)
		colors = append(colors,
			tablewriter.Colors{cpuColor},
			tablewriter.Colors{memColor},
		)
	}

	if wide {
		colors = append(colors,
			tablewriter.Colors{tablewriter.FgHiBlackColor}, // IP
			tablewriter.Colors{tablewriter.FgHiBlackColor}, // OS
			tablewriter.Colors{tablewriter.FgHiBlackColor}, // Kernel
		)
	}

	return colors
}

func getResourceColor(percent float64) int {
	switch {
	case percent > 90:
		return tablewriter.FgRedColor
	case percent > 70:
		return tablewriter.FgYellowColor
	default:
		return tablewriter.FgGreenColor
	}
}

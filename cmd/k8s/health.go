package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/beheshti/devops-toolkit/pkg/k8s"
	"github.com/beheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Check cluster health status",
		Long: `Comprehensive health check of your Kubernetes cluster.

Checks:
  • Node status and resource utilization
  • Pod health across namespaces
  • PersistentVolumeClaim status
  • Certificate expiration
  • Component status
  • Recent warning events`,
		RunE: runHealth,
	}

	cmd.Flags().Bool("watch", false, "Watch for changes")
	cmd.Flags().Duration("interval", 5*time.Second, "Watch interval")

	return cmd
}

func runHealth(cmd *cobra.Command, args []string) error {
	output.StartSpinner("Connecting to cluster...")

	client, err := k8s.NewClient(
		cmd.Flag("kubeconfig").Value.String(),
		cmd.Flag("context").Value.String(),
	)
	if err != nil {
		output.SpinnerError("Failed to connect to cluster")
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx := context.Background()

	output.SpinnerSuccess("Connected to cluster")
	output.Newline()

	// Get cluster info
	clusterInfo, err := client.GetClusterInfo(ctx)
	if err != nil {
		output.Warning("Could not get cluster info: " + err.Error())
	} else {
		output.Header(fmt.Sprintf("Cluster: %s", clusterInfo.Name))
	}

	// Create health summary table
	healthTable := output.NewTable(output.TableConfig{
		Title:      "Cluster Health Summary",
		Headers:    []string{"Component", "Status", "Details"},
		ShowBorder: true,
	})

	// Check nodes
	output.StartSpinner("Checking nodes...")
	nodeHealth, err := client.GetNodeHealth(ctx)
	if err != nil {
		output.SpinnerError("Failed to check nodes")
	} else {
		output.StopSpinner()
		status := fmt.Sprintf("%s %d/%d Ready", getStatusIcon(nodeHealth.Healthy), nodeHealth.Ready, nodeHealth.Total)
		row, colors := output.StatusRow("Nodes", getHealthStatus(nodeHealth.Healthy), status)
		healthTable.AddColoredRow(row, colors)
	}

	// Check pods
	output.StartSpinner("Checking pods...")
	namespace := cmd.Flag("namespace").Value.String()
	podHealth, err := client.GetPodHealth(ctx, namespace)
	if err != nil {
		output.SpinnerError("Failed to check pods")
	} else {
		output.StopSpinner()
		details := fmt.Sprintf("Running: %d, Pending: %d, Failed: %d",
			podHealth.Running, podHealth.Pending, podHealth.Failed)
		var status string
		if podHealth.Failed > 0 {
			status = fmt.Sprintf("%s %d Failed", output.IconError, podHealth.Failed)
		} else if podHealth.Pending > 5 {
			status = fmt.Sprintf("%s %d Pending", output.IconWarning, podHealth.Pending)
		} else {
			status = fmt.Sprintf("%s Healthy", output.IconSuccess)
		}
		row, colors := output.StatusRow("Pods", status, details)
		healthTable.AddColoredRow(row, colors)
	}

	// Check PVCs
	output.StartSpinner("Checking persistent volumes...")
	pvcHealth, err := client.GetPVCHealth(ctx, namespace)
	if err != nil {
		output.SpinnerError("Failed to check PVCs")
	} else {
		output.StopSpinner()
		healthy := pvcHealth.Pending == 0
		details := fmt.Sprintf("Bound: %d, Pending: %d", pvcHealth.Bound, pvcHealth.Pending)
		status := fmt.Sprintf("%s %s", getStatusIcon(healthy), getHealthStatus(healthy))
		row, colors := output.StatusRow("PVCs", status, details)
		healthTable.AddColoredRow(row, colors)
	}

	// Check deployments
	output.StartSpinner("Checking deployments...")
	deployHealth, err := client.GetDeploymentHealth(ctx, namespace)
	if err != nil {
		output.SpinnerError("Failed to check deployments")
	} else {
		output.StopSpinner()
		healthy := deployHealth.Unavailable == 0
		details := fmt.Sprintf("Ready: %d/%d, Unavailable: %d",
			deployHealth.Ready, deployHealth.Total, deployHealth.Unavailable)
		status := fmt.Sprintf("%s %s", getStatusIcon(healthy), getHealthStatus(healthy))
		row, colors := output.StatusRow("Deployments", status, details)
		healthTable.AddColoredRow(row, colors)
	}

	// Check services
	output.StartSpinner("Checking services...")
	svcHealth, err := client.GetServiceHealth(ctx, namespace)
	if err != nil {
		output.SpinnerError("Failed to check services")
	} else {
		output.StopSpinner()
		details := fmt.Sprintf("ClusterIP: %d, LoadBalancer: %d, NodePort: %d",
			svcHealth.ClusterIP, svcHealth.LoadBalancer, svcHealth.NodePort)
		row, colors := output.StatusRow("Services", fmt.Sprintf("%s OK", output.IconSuccess), details)
		healthTable.AddColoredRow(row, colors)
	}

	output.Newline()
	healthTable.Render()

	// Resource utilization
	output.Newline()
	output.StartSpinner("Getting resource utilization...")
	resources, err := client.GetResourceUtilization(ctx)
	if err != nil {
		output.SpinnerError("Could not get resource utilization (metrics-server may not be installed)")
	} else {
		output.StopSpinner()
		
		resourceTable := output.NewTable(output.TableConfig{
			Title:      "Resource Utilization",
			Headers:    []string{"Resource", "Used", "Capacity", "Utilization"},
			ShowBorder: true,
		})

		cpuUtil := float64(resources.CPUUsed) / float64(resources.CPUCapacity) * 100
		memUtil := float64(resources.MemoryUsed) / float64(resources.MemoryCapacity) * 100

		cpuBar := output.ProgressBar(int(cpuUtil), 100, 20)
		memBar := output.ProgressBar(int(memUtil), 100, 20)

		resourceTable.AddColoredRow(
			[]string{"CPU", fmt.Sprintf("%dm", resources.CPUUsed), fmt.Sprintf("%dm", resources.CPUCapacity), cpuBar},
			getUtilColors(cpuUtil),
		)
		resourceTable.AddColoredRow(
			[]string{"Memory", formatBytes(resources.MemoryUsed), formatBytes(resources.MemoryCapacity), memBar},
			getUtilColors(memUtil),
		)

		output.Newline()
		resourceTable.Render()
	}

	// Recent warning events
	output.Newline()
	output.StartSpinner("Getting recent events...")
	events, err := client.GetWarningEvents(ctx, namespace, 10)
	if err != nil {
		output.SpinnerError("Failed to get events")
	} else {
		output.StopSpinner()
		if len(events) > 0 {
			eventTable := output.NewTable(output.TableConfig{
				Title:      "Recent Warning Events",
				Headers:    []string{"Age", "Type", "Object", "Reason", "Message"},
				ShowBorder: true,
			})

			for _, event := range events {
				age := formatAge(event.LastTimestamp)
				eventTable.AddColoredRow(
					[]string{age, event.Type, event.Object, event.Reason, truncate(event.Message, 50)},
					[]tablewriter.Colors{
						{tablewriter.FgHiBlackColor},
						{tablewriter.FgYellowColor},
						{tablewriter.FgCyanColor},
						{tablewriter.FgYellowColor},
						{tablewriter.FgWhiteColor},
					},
				)
			}

			output.Newline()
			eventTable.Render()
		} else {
			output.Newline()
			output.Success("No warning events in the last hour")
		}
	}

	output.Newline()
	return nil
}

func getStatusIcon(healthy bool) string {
	if healthy {
		return output.IconSuccess
	}
	return output.IconError
}

func getHealthStatus(healthy bool) string {
	if healthy {
		return "Healthy"
	}
	return "Unhealthy"
}

func getUtilColors(util float64) []tablewriter.Colors {
	var statusColor int
	switch {
	case util > 90:
		statusColor = tablewriter.FgRedColor
	case util > 70:
		statusColor = tablewriter.FgYellowColor
	default:
		statusColor = tablewriter.FgGreenColor
	}

	return []tablewriter.Colors{
		{tablewriter.FgCyanColor},
		{tablewriter.FgWhiteColor},
		{tablewriter.FgHiBlackColor},
		{statusColor},
	}
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGi", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMi", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKi", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func formatAge(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}


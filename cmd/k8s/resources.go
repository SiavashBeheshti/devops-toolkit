package k8s

import (
	"context"
	"fmt"

	"github.com/SiavashBeheshti/devops-toolkit/pkg/k8s"
	"github.com/SiavashBeheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newResourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resources",
		Aliases: []string{"res", "usage"},
		Short:   "Show resource usage",
		Long: `Display resource usage across the cluster or namespace.

Shows:
  • CPU and Memory requests vs limits
  • Actual usage (requires metrics-server)
  • Over-provisioned resources
  • Resource quotas`,
		RunE: runResources,
	}

	cmd.Flags().Bool("top-pods", false, "Show top resource consuming pods")
	cmd.Flags().Int("limit", 10, "Number of top pods to show")

	return cmd
}

func runResources(cmd *cobra.Command, args []string) error {
	output.StartSpinner("Fetching resource data...")

	client, err := k8s.NewClient(
		cmd.Flag("kubeconfig").Value.String(),
		cmd.Flag("context").Value.String(),
	)
	if err != nil {
		output.SpinnerError("Failed to connect to cluster")
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx := context.Background()
	namespace := cmd.Flag("namespace").Value.String()
	showTopPods, _ := cmd.Flags().GetBool("top-pods")
	limit, _ := cmd.Flags().GetInt("limit")

	output.StopSpinner()
	output.Header("Resource Usage")

	// Cluster-wide resource summary
	output.StartSpinner("Getting cluster resources...")
	clusterRes, err := client.GetClusterResources(ctx)
	if err != nil {
		output.SpinnerError("Failed to get cluster resources")
		return err
	}
	output.StopSpinner()

	// Cluster summary table
	summaryTable := output.NewTable(output.TableConfig{
		Title:      "Cluster Resource Summary",
		Headers:    []string{"Resource", "Requests", "Limits", "Allocatable", "Utilization"},
		ShowBorder: true,
	})

	cpuReqPercent := float64(clusterRes.CPURequests) / float64(clusterRes.CPUAllocatable) * 100
	cpuLimPercent := float64(clusterRes.CPULimits) / float64(clusterRes.CPUAllocatable) * 100
	memReqPercent := float64(clusterRes.MemoryRequests) / float64(clusterRes.MemoryAllocatable) * 100
	memLimPercent := float64(clusterRes.MemoryLimits) / float64(clusterRes.MemoryAllocatable) * 100

	summaryTable.AddColoredRow(
		[]string{
			"CPU",
			fmt.Sprintf("%dm (%.1f%%)", clusterRes.CPURequests, cpuReqPercent),
			fmt.Sprintf("%dm (%.1f%%)", clusterRes.CPULimits, cpuLimPercent),
			fmt.Sprintf("%dm", clusterRes.CPUAllocatable),
			output.ProgressBar(int(cpuReqPercent), 100, 20),
		},
		getResourceRowColors(cpuReqPercent),
	)

	summaryTable.AddColoredRow(
		[]string{
			"Memory",
			fmt.Sprintf("%s (%.1f%%)", formatBytes(clusterRes.MemoryRequests), memReqPercent),
			fmt.Sprintf("%s (%.1f%%)", formatBytes(clusterRes.MemoryLimits), memLimPercent),
			formatBytes(clusterRes.MemoryAllocatable),
			output.ProgressBar(int(memReqPercent), 100, 20),
		},
		getResourceRowColors(memReqPercent),
	)

	summaryTable.AddColoredRow(
		[]string{
			"Pods",
			fmt.Sprintf("%d", clusterRes.PodCount),
			"-",
			fmt.Sprintf("%d", clusterRes.PodCapacity),
			output.ProgressBar(clusterRes.PodCount*100/clusterRes.PodCapacity, 100, 20),
		},
		getResourceRowColors(float64(clusterRes.PodCount)/float64(clusterRes.PodCapacity)*100),
	)

	summaryTable.Render()

	// Namespace breakdown if all namespaces
	if namespace == "" {
		output.Newline()
		output.StartSpinner("Getting namespace breakdown...")
		nsResources, err := client.GetNamespaceResources(ctx)
		if err != nil {
			output.SpinnerError("Failed to get namespace resources")
		} else {
			output.StopSpinner()

			nsTable := output.NewTable(output.TableConfig{
				Title:      "Resource Usage by Namespace",
				Headers:    []string{"Namespace", "Pods", "CPU Requests", "Memory Requests", "CPU %", "Mem %"},
				ShowBorder: true,
			})

			for _, ns := range nsResources {
				cpuPercent := float64(ns.CPURequests) / float64(clusterRes.CPUAllocatable) * 100
				memPercent := float64(ns.MemoryRequests) / float64(clusterRes.MemoryAllocatable) * 100

				nsTable.AddColoredRow(
					[]string{
						ns.Namespace,
						fmt.Sprintf("%d", ns.PodCount),
						fmt.Sprintf("%dm", ns.CPURequests),
						formatBytes(ns.MemoryRequests),
						fmt.Sprintf("%.1f%%", cpuPercent),
						fmt.Sprintf("%.1f%%", memPercent),
					},
					[]tablewriter.Colors{
						{tablewriter.FgCyanColor},
						{tablewriter.FgWhiteColor},
						{tablewriter.FgWhiteColor},
						{tablewriter.FgWhiteColor},
						{getResourceColorInt(cpuPercent)},
						{getResourceColorInt(memPercent)},
					},
				)
			}

			output.Newline()
			nsTable.Render()
		}
	}

	// Top resource consuming pods
	if showTopPods {
		output.Newline()
		output.StartSpinner("Getting top pods...")
		topPods, err := client.GetTopPods(ctx, namespace, limit)
		if err != nil {
			output.SpinnerError("Failed to get top pods (metrics-server required)")
		} else {
			output.StopSpinner()

			// CPU top
			cpuTable := output.NewTable(output.TableConfig{
				Title:      "Top Pods by CPU",
				Headers:    []string{"#", "Namespace", "Pod", "CPU Usage", "CPU Request", "Utilization"},
				ShowBorder: true,
			})

			for i, pod := range topPods.ByCPU {
				utilPercent := 0.0
				if pod.CPURequest > 0 {
					utilPercent = float64(pod.CPUUsage) / float64(pod.CPURequest) * 100
				}
				cpuTable.AddColoredRow(
					[]string{
						fmt.Sprintf("%d", i+1),
						pod.Namespace,
						pod.Name,
						fmt.Sprintf("%dm", pod.CPUUsage),
						fmt.Sprintf("%dm", pod.CPURequest),
						output.ProgressBar(int(utilPercent), 100, 15),
					},
					[]tablewriter.Colors{
						{tablewriter.FgHiBlackColor},
						{tablewriter.FgCyanColor},
						{tablewriter.FgWhiteColor},
						{tablewriter.FgYellowColor},
						{tablewriter.FgHiBlackColor},
						{getResourceColorInt(utilPercent)},
					},
				)
			}

			output.Newline()
			cpuTable.Render()

			// Memory top
			memTable := output.NewTable(output.TableConfig{
				Title:      "Top Pods by Memory",
				Headers:    []string{"#", "Namespace", "Pod", "Mem Usage", "Mem Request", "Utilization"},
				ShowBorder: true,
			})

			for i, pod := range topPods.ByMemory {
				utilPercent := 0.0
				if pod.MemoryRequest > 0 {
					utilPercent = float64(pod.MemoryUsage) / float64(pod.MemoryRequest) * 100
				}
				memTable.AddColoredRow(
					[]string{
						fmt.Sprintf("%d", i+1),
						pod.Namespace,
						pod.Name,
						formatBytes(pod.MemoryUsage),
						formatBytes(pod.MemoryRequest),
						output.ProgressBar(int(utilPercent), 100, 15),
					},
					[]tablewriter.Colors{
						{tablewriter.FgHiBlackColor},
						{tablewriter.FgCyanColor},
						{tablewriter.FgWhiteColor},
						{tablewriter.FgYellowColor},
						{tablewriter.FgHiBlackColor},
						{getResourceColorInt(utilPercent)},
					},
				)
			}

			output.Newline()
			memTable.Render()
		}
	}

	output.Newline()
	return nil
}

func getResourceRowColors(percent float64) []tablewriter.Colors {
	color := getResourceColorInt(percent)
	return []tablewriter.Colors{
		{tablewriter.FgCyanColor},
		{tablewriter.FgWhiteColor},
		{tablewriter.FgHiBlackColor},
		{tablewriter.FgHiBlackColor},
		{color},
	}
}

func getResourceColorInt(percent float64) int {
	switch {
	case percent > 90:
		return tablewriter.FgRedColor
	case percent > 70:
		return tablewriter.FgYellowColor
	default:
		return tablewriter.FgGreenColor
	}
}

package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/beheshti/devops-toolkit/pkg/k8s"
	"github.com/beheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newPodsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pods",
		Short: "List and analyze pods",
		Long: `List pods with enhanced visibility and analysis.

Features:
  • Color-coded status indicators
  • Resource usage display
  • Restart count highlighting
  • Age formatting
  • Grouping by status`,
		RunE: runPods,
	}

	cmd.Flags().BoolP("all-namespaces", "A", false, "List pods in all namespaces")
	cmd.Flags().Bool("problems", false, "Show only problematic pods")
	cmd.Flags().Bool("wide", false, "Show additional information")
	cmd.Flags().StringP("sort", "s", "name", "Sort by: name, status, age, restarts, namespace")
	cmd.Flags().StringP("label", "l", "", "Label selector")

	return cmd
}

func runPods(cmd *cobra.Command, args []string) error {
	output.StartSpinner("Fetching pods...")

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
	allNamespaces, _ := cmd.Flags().GetBool("all-namespaces")
	problemsOnly, _ := cmd.Flags().GetBool("problems")
	wide, _ := cmd.Flags().GetBool("wide")
	sortBy, _ := cmd.Flags().GetString("sort")
	labelSelector, _ := cmd.Flags().GetString("label")

	if allNamespaces {
		namespace = ""
	}

	pods, err := client.ListPods(ctx, namespace, labelSelector)
	if err != nil {
		output.SpinnerError("Failed to fetch pods")
		return fmt.Errorf("failed to list pods: %w", err)
	}

	output.SpinnerSuccess(fmt.Sprintf("Found %d pods", len(pods)))
	output.Newline()

	// Filter problematic pods if requested
	if problemsOnly {
		var filtered []k8s.PodInfo
		for _, pod := range pods {
			if isProblemPod(pod) {
				filtered = append(filtered, pod)
			}
		}
		pods = filtered
		if len(pods) == 0 {
			output.Success("No problematic pods found!")
			return nil
		}
		output.Warning(fmt.Sprintf("Found %d problematic pods", len(pods)))
		output.Newline()
	}

	// Sort pods
	sortPods(pods, sortBy)

	// Build table
	headers := []string{"Namespace", "Name", "Ready", "Status", "Restarts", "Age"}
	if wide {
		headers = append(headers, "Node", "IP")
	}

	table := output.NewTable(output.TableConfig{
		Title:      "Pods",
		Headers:    headers,
		ShowBorder: true,
	})

	// Status summary
	statusCounts := make(map[string]int)
	for _, pod := range pods {
		statusCounts[pod.Status]++
	}

	for _, pod := range pods {
		ready := fmt.Sprintf("%d/%d", pod.ReadyContainers, pod.TotalContainers)
		restarts := fmt.Sprintf("%d", pod.Restarts)
		age := formatAge(pod.CreationTime)

		row := []string{pod.Namespace, pod.Name, ready, pod.Status, restarts, age}
		if wide {
			row = append(row, pod.Node, pod.IP)
		}

		colors := getPodRowColors(pod, wide)
		table.AddColoredRow(row, colors)
	}

	table.Render()

	// Print summary
	output.Newline()
	printPodSummary(statusCounts)

	return nil
}

func isProblemPod(pod k8s.PodInfo) bool {
	problemStatuses := []string{
		"CrashLoopBackOff", "Error", "Failed", "ImagePullBackOff",
		"ErrImagePull", "Pending", "Unknown", "Terminating",
	}
	for _, status := range problemStatuses {
		if strings.Contains(pod.Status, status) {
			return true
		}
	}
	return pod.Restarts > 5 || pod.ReadyContainers < pod.TotalContainers
}

func sortPods(pods []k8s.PodInfo, sortBy string) {
	sort.Slice(pods, func(i, j int) bool {
		switch sortBy {
		case "status":
			return pods[i].Status < pods[j].Status
		case "age":
			return pods[i].CreationTime.After(pods[j].CreationTime)
		case "restarts":
			return pods[i].Restarts > pods[j].Restarts
		case "namespace":
			if pods[i].Namespace == pods[j].Namespace {
				return pods[i].Name < pods[j].Name
			}
			return pods[i].Namespace < pods[j].Namespace
		default: // name
			return pods[i].Name < pods[j].Name
		}
	})
}

func getPodRowColors(pod k8s.PodInfo, wide bool) []tablewriter.Colors {
	var statusColor int
	status := strings.ToLower(pod.Status)

	switch {
	case strings.Contains(status, "running"):
		statusColor = tablewriter.FgGreenColor
	case strings.Contains(status, "completed") || strings.Contains(status, "succeeded"):
		statusColor = tablewriter.FgCyanColor
	case strings.Contains(status, "pending"):
		statusColor = tablewriter.FgYellowColor
	case strings.Contains(status, "crash") || strings.Contains(status, "error") ||
		strings.Contains(status, "failed") || strings.Contains(status, "backoff"):
		statusColor = tablewriter.FgRedColor
	default:
		statusColor = tablewriter.FgWhiteColor
	}

	// Restart color
	var restartColor int
	switch {
	case pod.Restarts > 10:
		restartColor = tablewriter.FgRedColor
	case pod.Restarts > 3:
		restartColor = tablewriter.FgYellowColor
	default:
		restartColor = tablewriter.FgWhiteColor
	}

	// Ready color
	var readyColor int
	if pod.ReadyContainers < pod.TotalContainers {
		readyColor = tablewriter.FgYellowColor
	} else {
		readyColor = tablewriter.FgGreenColor
	}

	colors := []tablewriter.Colors{
		{tablewriter.FgCyanColor},    // namespace
		{tablewriter.FgWhiteColor},   // name
		{readyColor},                 // ready
		{tablewriter.Bold, statusColor}, // status
		{restartColor},               // restarts
		{tablewriter.FgHiBlackColor}, // age
	}

	if wide {
		colors = append(colors,
			tablewriter.Colors{tablewriter.FgHiBlackColor}, // node
			tablewriter.Colors{tablewriter.FgHiBlackColor}, // ip
		)
	}

	return colors
}

func printPodSummary(statusCounts map[string]int) {
	output.Print(output.Section("Summary"))

	running := statusCounts["Running"]
	pending := statusCounts["Pending"]
	failed := 0
	for status, count := range statusCounts {
		s := strings.ToLower(status)
		if strings.Contains(s, "error") || strings.Contains(s, "crash") ||
			strings.Contains(s, "failed") || strings.Contains(s, "backoff") {
			failed += count
		}
	}

	output.Printf("  %s Running: %d\n", output.SuccessStyle.Render(output.IconSuccess), running)
	if pending > 0 {
		output.Printf("  %s Pending: %d\n", output.WarningStyle.Render(output.IconWarning), pending)
	}
	if failed > 0 {
		output.Printf("  %s Failed/Crash: %d\n", output.ErrorStyle.Render(output.IconError), failed)
	}
	output.Newline()
}


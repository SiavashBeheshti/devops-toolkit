package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/SiavashBeheshti/devops-toolkit/pkg/k8s"
	"github.com/SiavashBeheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "Show cluster events",
		Long: `Display cluster events with filtering and highlighting.

Features:
  • Color-coded by event type
  • Filtering by type and object
  • Grouped by resource
  • Time-based filtering`,
		RunE: runEvents,
	}

	cmd.Flags().String("type", "", "Filter by event type (Normal, Warning)")
	cmd.Flags().String("reason", "", "Filter by reason")
	cmd.Flags().String("object", "", "Filter by object name")
	cmd.Flags().Int("limit", 50, "Maximum number of events to show")
	cmd.Flags().Bool("watch", false, "Watch for new events")
	cmd.Flags().Bool("warnings-only", false, "Show only warning events")

	return cmd
}

func runEvents(cmd *cobra.Command, args []string) error {
	output.StartSpinner("Fetching events...")

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
	eventType, _ := cmd.Flags().GetString("type")
	reason, _ := cmd.Flags().GetString("reason")
	objectFilter, _ := cmd.Flags().GetString("object")
	limit, _ := cmd.Flags().GetInt("limit")
	warningsOnly, _ := cmd.Flags().GetBool("warnings-only")

	if warningsOnly {
		eventType = "Warning"
	}

	events, err := client.ListEvents(ctx, namespace, k8s.EventFilter{
		Type:   eventType,
		Reason: reason,
		Object: objectFilter,
		Limit:  limit,
	})
	if err != nil {
		output.SpinnerError("Failed to fetch events")
		return fmt.Errorf("failed to list events: %w", err)
	}

	output.SpinnerSuccess(fmt.Sprintf("Found %d events", len(events)))
	output.Newline()

	if len(events) == 0 {
		output.Info("No events found matching the criteria")
		return nil
	}

	// Summary counts
	normalCount := 0
	warningCount := 0
	for _, e := range events {
		if e.Type == "Warning" {
			warningCount++
		} else {
			normalCount++
		}
	}

	// Event table
	table := output.NewTable(output.TableConfig{
		Title:      "Cluster Events",
		Headers:    []string{"Age", "Type", "Reason", "Object", "Message"},
		ShowBorder: true,
	})

	for _, event := range events {
		age := formatAge(event.LastTimestamp)
		object := fmt.Sprintf("%s/%s", strings.ToLower(event.Kind), event.Object)

		row := []string{
			age,
			event.Type,
			event.Reason,
			truncate(object, 40),
			truncate(event.Message, 60),
		}

		colors := getEventRowColors(event)
		table.AddColoredRow(row, colors)
	}

	table.Render()

	// Summary
	output.Newline()
	output.Print(output.Section("Event Summary"))
	output.Printf("  %s Normal: %d\n", output.SuccessStyle.Render(output.IconInfo), normalCount)
	if warningCount > 0 {
		output.Printf("  %s Warning: %d\n", output.WarningStyle.Render(output.IconWarning), warningCount)
	}

	// Group warnings by reason if there are many
	if warningCount > 5 {
		output.Newline()
		output.Print(output.SubSection("Warning Breakdown"))

		reasonCounts := make(map[string]int)
		for _, e := range events {
			if e.Type == "Warning" {
				reasonCounts[e.Reason]++
			}
		}

		for reason, count := range reasonCounts {
			output.Printf("    %s %s: %d\n",
				output.WarningStyle.Render(output.IconBullet),
				reason, count)
		}
	}

	output.Newline()
	return nil
}

func getEventRowColors(event k8s.EventInfo) []tablewriter.Colors {
	var typeColor int
	switch event.Type {
	case "Warning":
		typeColor = tablewriter.FgYellowColor
	case "Normal":
		typeColor = tablewriter.FgGreenColor
	default:
		typeColor = tablewriter.FgWhiteColor
	}

	// Reason color based on common problematic reasons
	var reasonColor int
	reason := strings.ToLower(event.Reason)
	switch {
	case strings.Contains(reason, "failed") || strings.Contains(reason, "error") ||
		strings.Contains(reason, "backoff") || strings.Contains(reason, "kill"):
		reasonColor = tablewriter.FgRedColor
	case strings.Contains(reason, "unhealthy") || strings.Contains(reason, "warning"):
		reasonColor = tablewriter.FgYellowColor
	case strings.Contains(reason, "created") || strings.Contains(reason, "started") ||
		strings.Contains(reason, "scheduled") || strings.Contains(reason, "pulled"):
		reasonColor = tablewriter.FgGreenColor
	default:
		reasonColor = tablewriter.FgCyanColor
	}

	return []tablewriter.Colors{
		{tablewriter.FgHiBlackColor},  // age
		{tablewriter.Bold, typeColor}, // type
		{reasonColor},                 // reason
		{tablewriter.FgCyanColor},     // object
		{tablewriter.FgWhiteColor},    // message
	}
}

package compliance

import (
	"github.com/SiavashBeheshti/devops-toolkit/pkg/compliance"
	"github.com/SiavashBeheshti/devops-toolkit/pkg/output"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func newPoliciesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "policies",
		Aliases: []string{"rules", "list"},
		Short:   "List available compliance policies",
		Long: `List all available compliance policies and rules.

Shows:
  • Built-in policies
  • Custom policies from policy directory
  • Rule details and severity`,
		RunE: runPolicies,
	}

	cmd.Flags().String("category", "", "Filter by category")
	cmd.Flags().String("severity", "", "Filter by severity")

	return cmd
}

func runPolicies(cmd *cobra.Command, args []string) error {
	category, _ := cmd.Flags().GetString("category")
	severity, _ := cmd.Flags().GetString("severity")

	output.Header("Compliance Policies")

	policies := compliance.GetBuiltinPolicies()

	// Filter
	if category != "" || severity != "" {
		var filtered []compliance.Policy
		for _, p := range policies {
			if category != "" && p.Category != category {
				continue
			}
			if severity != "" && p.Severity != severity {
				continue
			}
			filtered = append(filtered, p)
		}
		policies = filtered
	}

	if len(policies) == 0 {
		output.Info("No policies found matching the criteria")
		return nil
	}

	// Group by category
	byCategory := make(map[string][]compliance.Policy)
	for _, p := range policies {
		byCategory[p.Category] = append(byCategory[p.Category], p)
	}

	for cat, catPolicies := range byCategory {
		output.Newline()
		output.Print(output.Section(cat))

		table := output.NewTable(output.TableConfig{
			Headers:    []string{"ID", "Severity", "Name", "Description"},
			ShowBorder: true,
		})

		for _, p := range catPolicies {
			severityBadge := getSeverityBadge(p.Severity)

			table.AddColoredRow(
				[]string{
					p.ID,
					severityBadge,
					p.Name,
					truncateString(p.Description, 50),
				},
				getPolicyRowColors(p.Severity),
			)
		}

		table.Render()
	}

	// Summary
	output.Newline()
	output.Printf("Total: %d policies\n", len(policies))
	output.Newline()

	return nil
}

func getPolicyRowColors(severity string) []tablewriter.Colors {
	var severityColor int
	switch severity {
	case "critical":
		severityColor = tablewriter.FgRedColor
	case "high":
		severityColor = tablewriter.FgRedColor
	case "medium":
		severityColor = tablewriter.FgYellowColor
	default:
		severityColor = tablewriter.FgCyanColor
	}

	return []tablewriter.Colors{
		{tablewriter.FgCyanColor},    // ID
		{severityColor},              // Severity
		{tablewriter.FgWhiteColor},   // Name
		{tablewriter.FgHiBlackColor}, // Description
	}
}

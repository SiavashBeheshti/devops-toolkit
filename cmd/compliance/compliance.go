package compliance

import (
	"github.com/spf13/cobra"
)

// NewComplianceCmd creates the compliance command
func NewComplianceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "compliance",
		Aliases: []string{"comp", "policy"},
		Short:   "Compliance and security checking",
		Long: `Run compliance and security checks against your infrastructure.

Supports checking:
  • Kubernetes resources against best practices
  • Docker images for vulnerabilities
  • Infrastructure configurations
  • Custom policy rules`,
	}

	// Add subcommands
	cmd.AddCommand(newCheckCmd())
	cmd.AddCommand(newReportCmd())
	cmd.AddCommand(newPoliciesCmd())

	// Persistent flags
	cmd.PersistentFlags().StringP("policy-dir", "d", "", "Directory containing policy files")

	return cmd
}

package k8s

import (
	"github.com/beheshti/devops-toolkit/pkg/completion"
	"github.com/spf13/cobra"
)

// NewK8sCmd creates the k8s command
func NewK8sCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "k8s",
		Aliases: []string{"kubernetes", "kube"},
		Short:   "Kubernetes operations",
		Long: `Kubernetes operations for cluster management and debugging.

Commands provide enhanced visibility into your Kubernetes clusters
with beautiful, informative output that goes beyond kubectl.`,
	}

	// Add subcommands
	cmd.AddCommand(newHealthCmd())
	cmd.AddCommand(newPodsCmd())
	cmd.AddCommand(newNodesCmd())
	cmd.AddCommand(newCleanupCmd())
	cmd.AddCommand(newResourcesCmd())
	cmd.AddCommand(newEventsCmd())

	// Persistent flags for k8s commands
	cmd.PersistentFlags().StringP("namespace", "n", "", "Kubernetes namespace (default: all namespaces)")
	cmd.PersistentFlags().StringP("context", "c", "", "Kubernetes context to use")
	cmd.PersistentFlags().String("kubeconfig", "", "Path to kubeconfig file")

	// Register flag completions
	_ = cmd.RegisterFlagCompletionFunc("namespace", completion.NamespaceCompletion)
	_ = cmd.RegisterFlagCompletionFunc("context", completion.ContextCompletion)

	return cmd
}


package k8s

import (
	"context"
	"fmt"

	"github.com/SiavashBeheshti/devops-toolkit/pkg/k8s"
	"github.com/SiavashBeheshti/devops-toolkit/pkg/output"
	"github.com/spf13/cobra"
)

func newCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up cluster resources",
		Long: `Clean up unused or failed resources in the cluster.

Cleanup targets:
  • Completed/Failed pods
  • Evicted pods
  • Orphaned ReplicaSets
  • Completed Jobs
  • Unused ConfigMaps/Secrets (optional)`,
		RunE: runCleanup,
	}

	cmd.Flags().Bool("dry-run", true, "Show what would be deleted without deleting")
	cmd.Flags().Bool("completed-pods", true, "Clean up completed pods")
	cmd.Flags().Bool("failed-pods", true, "Clean up failed pods")
	cmd.Flags().Bool("evicted-pods", true, "Clean up evicted pods")
	cmd.Flags().Bool("completed-jobs", true, "Clean up completed jobs")
	cmd.Flags().Bool("orphan-rs", false, "Clean up orphaned ReplicaSets")
	cmd.Flags().Bool("force", false, "Skip confirmation")

	return cmd
}

func runCleanup(cmd *cobra.Command, args []string) error {
	output.StartSpinner("Analyzing cluster resources...")

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
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	cleanCompleted, _ := cmd.Flags().GetBool("completed-pods")
	cleanFailed, _ := cmd.Flags().GetBool("failed-pods")
	cleanEvicted, _ := cmd.Flags().GetBool("evicted-pods")
	cleanJobs, _ := cmd.Flags().GetBool("completed-jobs")
	cleanOrphanRS, _ := cmd.Flags().GetBool("orphan-rs")

	output.StopSpinner()
	output.Header("Cluster Cleanup")

	if dryRun {
		output.Info("Running in dry-run mode (no resources will be deleted)")
		output.Newline()
	}

	var totalCleaned int

	// Find and clean completed pods
	if cleanCompleted {
		output.StartSpinner("Finding completed pods...")
		pods, err := client.FindCompletedPods(ctx, namespace)
		if err != nil {
			output.SpinnerError("Failed to find completed pods")
		} else {
			output.StopSpinner()
			if len(pods) > 0 {
				output.Printf("\n%s Found %d completed pods:\n", output.InfoStyle.Render(output.IconInfo), len(pods))
				for _, pod := range pods {
					output.Printf("  %s %s/%s\n", output.MutedStyle.Render(output.IconBullet), pod.Namespace, pod.Name)
				}
				if !dryRun {
					deleted, err := client.DeletePods(ctx, pods)
					if err != nil {
						output.Error(fmt.Sprintf("Failed to delete some pods: %v", err))
					}
					totalCleaned += deleted
					output.Successf("Deleted %d completed pods", deleted)
				}
			} else {
				output.Success("No completed pods found")
			}
		}
	}

	// Find and clean failed pods
	if cleanFailed {
		output.StartSpinner("Finding failed pods...")
		pods, err := client.FindFailedPods(ctx, namespace)
		if err != nil {
			output.SpinnerError("Failed to find failed pods")
		} else {
			output.StopSpinner()
			if len(pods) > 0 {
				output.Printf("\n%s Found %d failed pods:\n", output.WarningStyle.Render(output.IconWarning), len(pods))
				for _, pod := range pods {
					output.Printf("  %s %s/%s (%s)\n",
						output.ErrorStyle.Render(output.IconBullet),
						pod.Namespace, pod.Name, pod.Status)
				}
				if !dryRun {
					deleted, err := client.DeletePods(ctx, pods)
					if err != nil {
						output.Error(fmt.Sprintf("Failed to delete some pods: %v", err))
					}
					totalCleaned += deleted
					output.Successf("Deleted %d failed pods", deleted)
				}
			} else {
				output.Success("No failed pods found")
			}
		}
	}

	// Find and clean evicted pods
	if cleanEvicted {
		output.StartSpinner("Finding evicted pods...")
		pods, err := client.FindEvictedPods(ctx, namespace)
		if err != nil {
			output.SpinnerError("Failed to find evicted pods")
		} else {
			output.StopSpinner()
			if len(pods) > 0 {
				output.Printf("\n%s Found %d evicted pods:\n", output.WarningStyle.Render(output.IconWarning), len(pods))
				for _, pod := range pods {
					output.Printf("  %s %s/%s\n",
						output.MutedStyle.Render(output.IconBullet),
						pod.Namespace, pod.Name)
				}
				if !dryRun {
					deleted, err := client.DeletePods(ctx, pods)
					if err != nil {
						output.Error(fmt.Sprintf("Failed to delete some pods: %v", err))
					}
					totalCleaned += deleted
					output.Successf("Deleted %d evicted pods", deleted)
				}
			} else {
				output.Success("No evicted pods found")
			}
		}
	}

	// Find and clean completed jobs
	if cleanJobs {
		output.StartSpinner("Finding completed jobs...")
		jobs, err := client.FindCompletedJobs(ctx, namespace)
		if err != nil {
			output.SpinnerError("Failed to find completed jobs")
		} else {
			output.StopSpinner()
			if len(jobs) > 0 {
				output.Printf("\n%s Found %d completed jobs:\n", output.InfoStyle.Render(output.IconInfo), len(jobs))
				for _, job := range jobs {
					output.Printf("  %s %s/%s\n",
						output.MutedStyle.Render(output.IconBullet),
						job.Namespace, job.Name)
				}
				if !dryRun {
					deleted, err := client.DeleteJobs(ctx, jobs)
					if err != nil {
						output.Error(fmt.Sprintf("Failed to delete some jobs: %v", err))
					}
					totalCleaned += deleted
					output.Successf("Deleted %d completed jobs", deleted)
				}
			} else {
				output.Success("No completed jobs found")
			}
		}
	}

	// Find and clean orphaned ReplicaSets
	if cleanOrphanRS {
		output.StartSpinner("Finding orphaned ReplicaSets...")
		replicaSets, err := client.FindOrphanedReplicaSets(ctx, namespace)
		if err != nil {
			output.SpinnerError("Failed to find orphaned ReplicaSets")
		} else {
			output.StopSpinner()
			if len(replicaSets) > 0 {
				output.Printf("\n%s Found %d orphaned ReplicaSets:\n", output.InfoStyle.Render(output.IconInfo), len(replicaSets))
				for _, rs := range replicaSets {
					output.Printf("  %s %s/%s\n",
						output.MutedStyle.Render(output.IconBullet),
						rs.Namespace, rs.Name)
				}
				if !dryRun {
					deleted, err := client.DeleteReplicaSets(ctx, replicaSets)
					if err != nil {
						output.Error(fmt.Sprintf("Failed to delete some ReplicaSets: %v", err))
					}
					totalCleaned += deleted
					output.Successf("Deleted %d orphaned ReplicaSets", deleted)
				}
			} else {
				output.Success("No orphaned ReplicaSets found")
			}
		}
	}

	// Summary
	output.Newline()
	output.Print(output.Divider(50))
	output.Newline()

	if dryRun {
		output.Info("Dry-run complete. Use --dry-run=false to actually delete resources.")
	} else {
		output.Successf("Cleanup complete! Removed %d resources.", totalCleaned)
	}

	output.Newline()
	return nil
}

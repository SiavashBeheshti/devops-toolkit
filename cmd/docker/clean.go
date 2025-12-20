package docker

import (
	"context"
	"fmt"

	"github.com/SiavashBeheshti/devops-toolkit/pkg/docker"
	"github.com/SiavashBeheshti/devops-toolkit/pkg/output"
	"github.com/spf13/cobra"
)

func newCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean up Docker resources",
		Long: `Clean up unused Docker resources to reclaim disk space.

Cleanup targets:
  • Stopped containers
  • Dangling images
  • Unused networks
  • Build cache
  • Unused volumes (with --volumes flag)`,
		RunE: runClean,
	}

	cmd.Flags().Bool("dry-run", true, "Show what would be deleted without deleting")
	cmd.Flags().Bool("containers", true, "Remove stopped containers")
	cmd.Flags().Bool("images", true, "Remove dangling images")
	cmd.Flags().Bool("networks", true, "Remove unused networks")
	cmd.Flags().Bool("volumes", false, "Remove unused volumes (dangerous!)")
	cmd.Flags().Bool("build-cache", true, "Remove build cache")
	cmd.Flags().Bool("all-images", false, "Remove all unused images (not just dangling)")
	cmd.Flags().Bool("force", false, "Skip confirmation")

	return cmd
}

func runClean(cmd *cobra.Command, args []string) error {
	output.StartSpinner("Analyzing Docker resources...")

	client, err := docker.NewClient()
	if err != nil {
		output.SpinnerError("Failed to connect to Docker")
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	cleanContainers, _ := cmd.Flags().GetBool("containers")
	cleanImages, _ := cmd.Flags().GetBool("images")
	cleanNetworks, _ := cmd.Flags().GetBool("networks")
	cleanVolumes, _ := cmd.Flags().GetBool("volumes")
	cleanBuildCache, _ := cmd.Flags().GetBool("build-cache")
	allImages, _ := cmd.Flags().GetBool("all-images")

	output.StopSpinner()
	output.Header("Docker Cleanup")

	if dryRun {
		output.Info("Running in dry-run mode (no resources will be deleted)")
		output.Newline()
	}

	var totalSpaceReclaimed int64

	// Clean stopped containers
	if cleanContainers {
		output.StartSpinner("Finding stopped containers...")
		containers, err := client.FindStoppedContainers(ctx)
		if err != nil {
			output.SpinnerError("Failed to find containers")
		} else {
			output.StopSpinner()
			if len(containers) > 0 {
				output.Printf("\n%s Found %d stopped containers:\n",
					output.InfoStyle.Render(output.IconInfo), len(containers))
				for _, c := range containers {
					output.Printf("  %s %s (%s)\n",
						output.MutedStyle.Render(output.IconBullet),
						c.Name, truncateID(c.ID))
				}
				if !dryRun {
					deleted, space, err := client.RemoveContainers(ctx, containers)
					if err != nil {
						output.Error(fmt.Sprintf("Failed to remove some containers: %v", err))
					}
					totalSpaceReclaimed += space
					output.Successf("Removed %d containers", deleted)
				}
			} else {
				output.Success("No stopped containers found")
			}
		}
	}

	// Clean images
	if cleanImages {
		output.StartSpinner("Finding unused images...")
		images, err := client.FindUnusedImages(ctx, allImages)
		if err != nil {
			output.SpinnerError("Failed to find images")
		} else {
			output.StopSpinner()
			if len(images) > 0 {
				var totalSize int64
				for _, img := range images {
					totalSize += img.Size
				}

				label := "dangling"
				if allImages {
					label = "unused"
				}

				output.Printf("\n%s Found %d %s images (%s):\n",
					output.InfoStyle.Render(output.IconInfo),
					len(images), label, formatSize(totalSize))

				for _, img := range images {
					name := img.Repository
					if img.Tag != "" && img.Tag != "<none>" {
						name = fmt.Sprintf("%s:%s", img.Repository, img.Tag)
					}
					output.Printf("  %s %s (%s)\n",
						output.MutedStyle.Render(output.IconBullet),
						name, formatSize(img.Size))
				}

				if !dryRun {
					deleted, space, err := client.RemoveImages(ctx, images)
					if err != nil {
						output.Error(fmt.Sprintf("Failed to remove some images: %v", err))
					}
					totalSpaceReclaimed += space
					output.Successf("Removed %d images, reclaimed %s", deleted, formatSize(space))
				}
			} else {
				output.Success("No unused images found")
			}
		}
	}

	// Clean networks
	if cleanNetworks {
		output.StartSpinner("Finding unused networks...")
		networks, err := client.FindUnusedNetworks(ctx)
		if err != nil {
			output.SpinnerError("Failed to find networks")
		} else {
			output.StopSpinner()
			if len(networks) > 0 {
				output.Printf("\n%s Found %d unused networks:\n",
					output.InfoStyle.Render(output.IconInfo), len(networks))
				for _, n := range networks {
					output.Printf("  %s %s\n",
						output.MutedStyle.Render(output.IconBullet), n.Name)
				}
				if !dryRun {
					deleted, err := client.RemoveNetworks(ctx, networks)
					if err != nil {
						output.Error(fmt.Sprintf("Failed to remove some networks: %v", err))
					}
					output.Successf("Removed %d networks", deleted)
				}
			} else {
				output.Success("No unused networks found")
			}
		}
	}

	// Clean volumes (dangerous!)
	if cleanVolumes {
		output.StartSpinner("Finding unused volumes...")
		volumes, err := client.FindUnusedVolumes(ctx)
		if err != nil {
			output.SpinnerError("Failed to find volumes")
		} else {
			output.StopSpinner()
			if len(volumes) > 0 {
				var totalSize int64
				for _, v := range volumes {
					totalSize += v.Size
				}

				output.Printf("\n%s Found %d unused volumes (%s):\n",
					output.WarningStyle.Render(output.IconWarning),
					len(volumes), formatSize(totalSize))

				for _, v := range volumes {
					output.Printf("  %s %s (%s)\n",
						output.WarningStyle.Render(output.IconBullet),
						v.Name, formatSize(v.Size))
				}

				if !dryRun {
					deleted, space, err := client.RemoveVolumes(ctx, volumes)
					if err != nil {
						output.Error(fmt.Sprintf("Failed to remove some volumes: %v", err))
					}
					totalSpaceReclaimed += space
					output.Successf("Removed %d volumes, reclaimed %s", deleted, formatSize(space))
				}
			} else {
				output.Success("No unused volumes found")
			}
		}
	}

	// Clean build cache
	if cleanBuildCache {
		output.StartSpinner("Analyzing build cache...")
		cacheSize, err := client.GetBuildCacheSize(ctx)
		if err != nil {
			output.SpinnerError("Failed to analyze build cache")
		} else {
			output.StopSpinner()
			if cacheSize > 0 {
				output.Printf("\n%s Build cache using %s\n",
					output.InfoStyle.Render(output.IconInfo), formatSize(cacheSize))

				if !dryRun {
					reclaimed, err := client.PruneBuildCache(ctx)
					if err != nil {
						output.Error(fmt.Sprintf("Failed to prune build cache: %v", err))
					} else {
						totalSpaceReclaimed += reclaimed
						output.Successf("Cleared build cache, reclaimed %s", formatSize(reclaimed))
					}
				}
			} else {
				output.Success("Build cache is empty")
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
		output.Successf("Cleanup complete! Reclaimed %s of disk space.", formatSize(totalSpaceReclaimed))
	}

	output.Newline()
	return nil
}

package completion

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

// getDockerClient creates a Docker client for completion
func getDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// ContainerCompletion provides Docker container name/ID completion
func ContainerCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cli, err := getDockerClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, c := range containers {
		// Complete by container ID (short)
		shortID := c.ID[:12]
		if strings.HasPrefix(shortID, toComplete) {
			completions = append(completions, shortID)
		}

		// Complete by container name
		for _, name := range c.Names {
			name = strings.TrimPrefix(name, "/")
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name)
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// RunningContainerCompletion provides completion for running Docker containers only
func RunningContainerCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cli, err := getDockerClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: false})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, c := range containers {
		// Complete by container ID (short)
		shortID := c.ID[:12]
		if strings.HasPrefix(shortID, toComplete) {
			completions = append(completions, shortID)
		}

		// Complete by container name
		for _, name := range c.Names {
			name = strings.TrimPrefix(name, "/")
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name)
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// ImageCompletion provides Docker image name/ID completion
func ImageCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cli, err := getDockerClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer cli.Close()

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{All: false})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	seen := make(map[string]bool)

	for _, img := range images {
		// Complete by image ID (short)
		shortID := strings.TrimPrefix(img.ID, "sha256:")[:12]
		if strings.HasPrefix(shortID, toComplete) && !seen[shortID] {
			completions = append(completions, shortID)
			seen[shortID] = true
		}

		// Complete by repo tags
		for _, tag := range img.RepoTags {
			if tag == "<none>:<none>" {
				continue
			}
			if strings.HasPrefix(tag, toComplete) && !seen[tag] {
				completions = append(completions, tag)
				seen[tag] = true
			}
			// Also match partial repo name
			parts := strings.Split(tag, ":")
			if len(parts) > 0 && strings.HasPrefix(parts[0], toComplete) && !seen[tag] {
				completions = append(completions, tag)
				seen[tag] = true
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// NetworkCompletion provides Docker network name completion
func NetworkCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cli, err := getDockerClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer cli.Close()

	networks, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, net := range networks {
		if strings.HasPrefix(net.Name, toComplete) {
			completions = append(completions, net.Name)
		}
		if strings.HasPrefix(net.ID[:12], toComplete) {
			completions = append(completions, net.ID[:12])
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// VolumeCompletion provides Docker volume name completion
func VolumeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cli, err := getDockerClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer cli.Close()

	volumes, err := cli.VolumeList(context.Background(), volume.ListOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, vol := range volumes.Volumes {
		if strings.HasPrefix(vol.Name, toComplete) {
			completions = append(completions, vol.Name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// LogLevelCompletion provides log level completion
func LogLevelCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	levels := []string{"error", "warn", "info", "debug"}

	var completions []string
	for _, level := range levels {
		if strings.HasPrefix(level, toComplete) {
			completions = append(completions, level)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}


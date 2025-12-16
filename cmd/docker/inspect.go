package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/beheshti/devops-toolkit/pkg/docker"
	"github.com/beheshti/devops-toolkit/pkg/output"
	"github.com/spf13/cobra"
)

func newInspectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect [container]",
		Short: "Inspect container details",
		Long: `Display detailed information about a container in a readable format.

Shows:
  • Container configuration
  • Network settings
  • Mount points
  • Environment variables
  • Health check status`,
		Args: cobra.ExactArgs(1),
		RunE: runInspect,
	}

	cmd.Flags().Bool("env", false, "Show environment variables")
	cmd.Flags().Bool("mounts", false, "Show mount details")
	cmd.Flags().Bool("network", false, "Show network details")
	cmd.Flags().Bool("all", false, "Show all details")

	return cmd
}

func runInspect(cmd *cobra.Command, args []string) error {
	containerID := args[0]

	output.StartSpinner(fmt.Sprintf("Inspecting container %s...", containerID))

	client, err := docker.NewClient()
	if err != nil {
		output.SpinnerError("Failed to connect to Docker")
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	showEnv, _ := cmd.Flags().GetBool("env")
	showMounts, _ := cmd.Flags().GetBool("mounts")
	showNetwork, _ := cmd.Flags().GetBool("network")
	showAll, _ := cmd.Flags().GetBool("all")

	if showAll {
		showEnv = true
		showMounts = true
		showNetwork = true
	}

	info, err := client.InspectContainer(ctx, containerID)
	if err != nil {
		output.SpinnerError("Failed to inspect container")
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	output.SpinnerSuccess("Container found")
	output.Newline()

	// Basic info
	output.Header(fmt.Sprintf("Container: %s", info.Name))

	output.Print(output.Section("Basic Information"))
	output.Printf("  %s\n", output.KeyValue("ID", info.ID))
	output.Printf("  %s\n", output.KeyValue("Image", info.Image))
	output.Printf("  %s\n", output.KeyValue("Status", formatStatus(info.State, info.Status)))
	output.Printf("  %s\n", output.KeyValue("Created", info.Created))
	output.Printf("  %s\n", output.KeyValue("Started", info.StartedAt))
	if info.FinishedAt != "" && info.State != "running" {
		output.Printf("  %s\n", output.KeyValue("Finished", info.FinishedAt))
	}
	output.Printf("  %s\n", output.KeyValue("Restart Count", fmt.Sprintf("%d", info.RestartCount)))
	output.Printf("  %s\n", output.KeyValue("Platform", info.Platform))

	// Health
	if info.Health != "" {
		output.Newline()
		output.Print(output.Section("Health Check"))
		healthIcon := output.StatusIcon(info.Health)
		output.Printf("  %s %s\n", healthIcon, info.Health)
		if info.HealthLog != "" {
			output.Printf("  Last check: %s\n", info.HealthLog)
		}
	}

	// Command
	output.Newline()
	output.Print(output.Section("Command"))
	output.Printf("  %s\n", info.Command)
	if info.Entrypoint != "" {
		output.Printf("  Entrypoint: %s\n", info.Entrypoint)
	}

	// Ports
	if len(info.Ports) > 0 {
		output.Newline()
		output.Print(output.Section("Port Mappings"))
		for _, port := range info.Ports {
			if port.PublicPort > 0 {
				output.Printf("  %s %s:%d → %d/%s\n",
					output.InfoStyle.Render(output.IconArrow),
					port.IP, port.PublicPort, port.PrivatePort, port.Type)
			} else {
				output.Printf("  %s %d/%s (not published)\n",
					output.MutedStyle.Render(output.IconBullet),
					port.PrivatePort, port.Type)
			}
		}
	}

	// Environment variables
	if showEnv && len(info.Env) > 0 {
		output.Newline()
		output.Print(output.Section("Environment Variables"))
		for _, env := range info.Env {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				// Mask sensitive values
				value := parts[1]
				if isSensitiveEnv(parts[0]) {
					value = "********"
				}
				output.Printf("  %s=%s\n",
					output.InfoStyle.Render(parts[0]),
					output.MutedStyle.Render(value))
			}
		}
	}

	// Mounts
	if showMounts && len(info.Mounts) > 0 {
		output.Newline()
		output.Print(output.Section("Mounts"))
		for _, mount := range info.Mounts {
			rw := "rw"
			if !mount.RW {
				rw = "ro"
			}
			output.Printf("  %s %s\n", output.InfoStyle.Render(mount.Type), mount.Name)
			output.Printf("      %s → %s (%s)\n",
				mount.Source, mount.Destination, rw)
		}
	}

	// Network
	if showNetwork && len(info.Networks) > 0 {
		output.Newline()
		output.Print(output.Section("Networks"))
		for name, net := range info.Networks {
			output.Printf("  %s %s\n", output.InfoStyle.Render(output.IconBullet), name)
			output.Printf("      IP Address: %s\n", net.IPAddress)
			output.Printf("      Gateway: %s\n", net.Gateway)
			output.Printf("      MAC Address: %s\n", net.MacAddress)
		}
	}

	// Labels
	if len(info.Labels) > 0 {
		output.Newline()
		output.Print(output.Section("Labels"))
		count := 0
		for key, value := range info.Labels {
			if count >= 10 {
				output.Printf("  ... and %d more labels\n", len(info.Labels)-10)
				break
			}
			output.Printf("  %s: %s\n",
				output.MutedStyle.Render(key),
				truncate(value, 50))
			count++
		}
	}

	output.Newline()
	return nil
}

func formatStatus(state, status string) string {
	icon := output.StatusIcon(state)
	return fmt.Sprintf("%s %s", icon, status)
}

func isSensitiveEnv(name string) bool {
	sensitive := []string{
		"PASSWORD", "SECRET", "KEY", "TOKEN", "CREDENTIAL",
		"API_KEY", "APIKEY", "AUTH", "PRIVATE",
	}
	upper := strings.ToUpper(name)
	for _, s := range sensitive {
		if strings.Contains(upper, s) {
			return true
		}
	}
	return false
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}


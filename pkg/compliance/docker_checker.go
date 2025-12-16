package compliance

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// DockerChecker checks Docker resources for compliance
type DockerChecker struct {
	opts   CheckOptions
	client *client.Client
}

// NewDockerChecker creates a new Docker checker
func NewDockerChecker(opts CheckOptions) *DockerChecker {
	return &DockerChecker{opts: opts}
}

// Run runs the Docker compliance checks
func (c *DockerChecker) Run(ctx context.Context) ([]CheckResult, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	c.client = cli
	defer cli.Close()

	var results []CheckResult

	// Container security checks
	containerResults, err := c.checkContainerSecurity(ctx)
	if err == nil {
		results = append(results, containerResults...)
	}

	// Image checks
	if c.opts.Image != "" {
		imageResults, err := c.checkImage(ctx, c.opts.Image)
		if err == nil {
			results = append(results, imageResults...)
		}
	}

	return results, nil
}

func (c *DockerChecker) checkContainerSecurity(ctx context.Context) ([]CheckResult, error) {
	var results []CheckResult

	containers, err := c.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	for _, cont := range containers {
		name := cont.Names[0]
		if strings.HasPrefix(name, "/") {
			name = name[1:]
		}

		// Inspect container for detailed info
		inspect, err := c.client.ContainerInspect(ctx, cont.ID)
		if err != nil {
			continue
		}

		// Check privileged mode
		if inspect.HostConfig.Privileged {
			results = append(results, CheckResult{
				RuleID:      "DOCKER-SEC-001",
				RuleName:    "No Privileged Containers",
				Category:    "Docker Security",
				Severity:    "critical",
				Status:      StatusFailed,
				Resource:    name,
				Message:     "Container is running in privileged mode",
				Remediation: "Remove --privileged flag",
			})
		} else {
			results = append(results, CheckResult{
				RuleID:   "DOCKER-SEC-001",
				RuleName: "No Privileged Containers",
				Category: "Docker Security",
				Severity: "critical",
				Status:   StatusPassed,
				Resource: name,
				Message:  "Container is not running in privileged mode",
			})
		}

		// Check user namespace
		if inspect.HostConfig.UsernsMode == "" || inspect.HostConfig.UsernsMode == "host" {
			// Check if running as root
			if inspect.Config.User == "" || inspect.Config.User == "root" || inspect.Config.User == "0" {
				results = append(results, CheckResult{
					RuleID:      "DOCKER-SEC-002",
					RuleName:    "Non-Root User",
					Category:    "Docker Security",
					Severity:    "high",
					Status:      StatusFailed,
					Resource:    name,
					Message:     "Container is running as root",
					Remediation: "Use USER directive in Dockerfile or --user flag",
				})
			}
		}

		// Check host network
		if inspect.HostConfig.NetworkMode == "host" {
			results = append(results, CheckResult{
				RuleID:      "DOCKER-SEC-003",
				RuleName:    "No Host Network",
				Category:    "Docker Security",
				Severity:    "high",
				Status:      StatusFailed,
				Resource:    name,
				Message:     "Container is using host network",
				Remediation: "Use bridge or custom network",
			})
		}

		// Check host PID
		if inspect.HostConfig.PidMode == "host" {
			results = append(results, CheckResult{
				RuleID:      "DOCKER-SEC-004",
				RuleName:    "No Host PID",
				Category:    "Docker Security",
				Severity:    "high",
				Status:      StatusFailed,
				Resource:    name,
				Message:     "Container is using host PID namespace",
				Remediation: "Remove --pid=host flag",
			})
		}

		// Check capabilities
		if len(inspect.HostConfig.CapAdd) > 0 {
			for _, cap := range inspect.HostConfig.CapAdd {
				if isDangerousCap(cap) {
					results = append(results, CheckResult{
						RuleID:      "DOCKER-SEC-005",
						RuleName:    "No Dangerous Capabilities",
						Category:    "Docker Security",
						Severity:    "high",
						Status:      StatusFailed,
						Resource:    name,
						Message:     fmt.Sprintf("Container has dangerous capability: %s", cap),
						Remediation: "Remove unnecessary capabilities",
					})
				}
			}
		}

		// Check memory limits
		if inspect.HostConfig.Memory == 0 {
			results = append(results, CheckResult{
				RuleID:      "DOCKER-RES-001",
				RuleName:    "Memory Limits",
				Category:    "Docker Resources",
				Severity:    "medium",
				Status:      StatusFailed,
				Resource:    name,
				Message:     "Container has no memory limit",
				Remediation: "Set --memory flag",
			})
		}

		// Check CPU limits
		if inspect.HostConfig.CPUQuota == 0 && inspect.HostConfig.NanoCPUs == 0 {
			results = append(results, CheckResult{
				RuleID:      "DOCKER-RES-002",
				RuleName:    "CPU Limits",
				Category:    "Docker Resources",
				Severity:    "low",
				Status:      StatusFailed,
				Resource:    name,
				Message:     "Container has no CPU limit",
				Remediation: "Set --cpus or --cpu-quota flag",
			})
		}

		// Check restart policy
		if inspect.HostConfig.RestartPolicy.Name == "" || inspect.HostConfig.RestartPolicy.Name == "no" {
			results = append(results, CheckResult{
				RuleID:      "DOCKER-CFG-001",
				RuleName:    "Restart Policy",
				Category:    "Docker Configuration",
				Severity:    "low",
				Status:      StatusFailed,
				Resource:    name,
				Message:     "Container has no restart policy",
				Remediation: "Set --restart=unless-stopped or similar",
			})
		}

		// Check health check
		if inspect.Config.Healthcheck == nil || len(inspect.Config.Healthcheck.Test) == 0 {
			results = append(results, CheckResult{
				RuleID:      "DOCKER-CFG-002",
				RuleName:    "Health Check",
				Category:    "Docker Configuration",
				Severity:    "medium",
				Status:      StatusFailed,
				Resource:    name,
				Message:     "Container has no health check",
				Remediation: "Add HEALTHCHECK in Dockerfile or --health-cmd flag",
			})
		}

		// Check read-only root filesystem
		if !inspect.HostConfig.ReadonlyRootfs {
			results = append(results, CheckResult{
				RuleID:      "DOCKER-SEC-006",
				RuleName:    "Read-Only Root Filesystem",
				Category:    "Docker Security",
				Severity:    "medium",
				Status:      StatusFailed,
				Resource:    name,
				Message:     "Container has writable root filesystem",
				Remediation: "Use --read-only flag",
			})
		}
	}

	return results, nil
}

func (c *DockerChecker) checkImage(ctx context.Context, imageName string) ([]CheckResult, error) {
	var results []CheckResult

	// Inspect image
	inspect, _, err := c.client.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return nil, err
	}

	resource := imageName

	// Check for latest tag
	for _, tag := range inspect.RepoTags {
		if strings.HasSuffix(tag, ":latest") {
			results = append(results, CheckResult{
				RuleID:      "DOCKER-IMG-001",
				RuleName:    "No Latest Tag",
				Category:    "Docker Images",
				Severity:    "medium",
				Status:      StatusFailed,
				Resource:    resource,
				Message:     "Image uses 'latest' tag",
				Remediation: "Use specific version tags",
			})
			break
		}
	}

	// Check image size
	sizeMB := inspect.Size / (1024 * 1024)
	if sizeMB > 1000 {
		results = append(results, CheckResult{
			RuleID:      "DOCKER-IMG-002",
			RuleName:    "Image Size",
			Category:    "Docker Images",
			Severity:    "low",
			Status:      StatusFailed,
			Resource:    resource,
			Message:     fmt.Sprintf("Image is large: %d MB", sizeMB),
			Remediation: "Use multi-stage builds or smaller base images",
		})
	}

	// Check for root user in image
	if inspect.Config.User == "" || inspect.Config.User == "root" || inspect.Config.User == "0" {
		results = append(results, CheckResult{
			RuleID:      "DOCKER-IMG-003",
			RuleName:    "Non-Root User in Image",
			Category:    "Docker Images",
			Severity:    "medium",
			Status:      StatusFailed,
			Resource:    resource,
			Message:     "Image runs as root by default",
			Remediation: "Add USER directive in Dockerfile",
		})
	}

	// Check exposed ports
	if len(inspect.Config.ExposedPorts) > 0 {
		for port := range inspect.Config.ExposedPorts {
			portNum := port.Int()
			if portNum < 1024 {
				results = append(results, CheckResult{
					RuleID:      "DOCKER-IMG-004",
					RuleName:    "Privileged Ports",
					Category:    "Docker Images",
					Severity:    "low",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Image exposes privileged port: %d", portNum),
					Remediation: "Use ports > 1024",
				})
			}
		}
	}

	return results, nil
}

func isDangerousCap(cap string) bool {
	dangerous := []string{
		"SYS_ADMIN",
		"SYS_PTRACE",
		"NET_ADMIN",
		"SYS_MODULE",
		"SYS_RAWIO",
		"SYS_BOOT",
		"MAC_ADMIN",
		"MAC_OVERRIDE",
	}

	for _, d := range dangerous {
		if cap == d {
			return true
		}
	}
	return false
}


package compliance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileChecker checks configuration files for compliance
type FileChecker struct {
	opts CheckOptions
}

// NewFileChecker creates a new file checker
func NewFileChecker(opts CheckOptions) *FileChecker {
	return &FileChecker{opts: opts}
}

// Run runs the file compliance checks
func (c *FileChecker) Run(ctx context.Context) ([]CheckResult, error) {
	var results []CheckResult

	// Walk through files
	err := filepath.Walk(c.opts.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Check Kubernetes manifests
		if isKubernetesManifest(path) {
			fileResults, err := c.checkKubernetesManifest(path)
			if err == nil {
				results = append(results, fileResults...)
			}
		}

		// Check Dockerfiles
		if isDockerfile(path) {
			fileResults, err := c.checkDockerfile(path)
			if err == nil {
				results = append(results, fileResults...)
			}
		}

		// Check docker-compose files
		if isDockerCompose(path) {
			fileResults, err := c.checkDockerCompose(path)
			if err == nil {
				results = append(results, fileResults...)
			}
		}

		return nil
	})

	return results, err
}

func isKubernetesManifest(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".yaml" && ext != ".yml" {
		return false
	}

	// Read first few lines to check for Kubernetes-like content
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	content := string(data)
	return strings.Contains(content, "apiVersion:") && strings.Contains(content, "kind:")
}

func isDockerfile(path string) bool {
	name := filepath.Base(path)
	return name == "Dockerfile" || strings.HasPrefix(name, "Dockerfile.")
}

func isDockerCompose(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return name == "docker-compose.yml" || name == "docker-compose.yaml" ||
		name == "compose.yml" || name == "compose.yaml"
}

func (c *FileChecker) checkKubernetesManifest(path string) ([]CheckResult, error) {
	var results []CheckResult
	resource := path

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manifest map[string]interface{}
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	kind, _ := manifest["kind"].(string)

	// Check for Deployment/Pod specific rules
	if kind == "Deployment" || kind == "Pod" || kind == "StatefulSet" || kind == "DaemonSet" {
		spec := getNestedMap(manifest, "spec")
		if spec == nil {
			return results, nil
		}

		// For Deployments, get pod template spec
		if kind != "Pod" {
			template := getNestedMap(spec, "template")
			if template != nil {
				spec = getNestedMap(template, "spec")
			}
		}

		if spec == nil {
			return results, nil
		}

		// Check containers
		containers, _ := spec["containers"].([]interface{})
		for _, c := range containers {
			container, _ := c.(map[string]interface{})
			containerName, _ := container["name"].(string)

			// Check image tag
			image, _ := container["image"].(string)
			if strings.HasSuffix(image, ":latest") || !strings.Contains(image, ":") {
				results = append(results, CheckResult{
					RuleID:      "FILE-K8S-001",
					RuleName:    "No Latest Tag",
					Category:    "File Compliance",
					Severity:    "medium",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' uses latest or no tag", containerName),
					Remediation: "Use specific image tags",
				})
			}

			// Check resources
			resources, _ := container["resources"].(map[string]interface{})
			if resources == nil {
				results = append(results, CheckResult{
					RuleID:      "FILE-K8S-002",
					RuleName:    "Resource Limits",
					Category:    "File Compliance",
					Severity:    "medium",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' has no resource limits", containerName),
					Remediation: "Add resources.limits",
				})
			} else {
				limits, _ := resources["limits"].(map[string]interface{})
				if limits == nil {
					results = append(results, CheckResult{
						RuleID:      "FILE-K8S-002",
						RuleName:    "Resource Limits",
						Category:    "File Compliance",
						Severity:    "medium",
						Status:      StatusFailed,
						Resource:    resource,
						Message:     fmt.Sprintf("Container '%s' has no resource limits", containerName),
						Remediation: "Add resources.limits",
					})
				}
			}

			// Check security context
			secContext, _ := container["securityContext"].(map[string]interface{})
			if secContext == nil {
				results = append(results, CheckResult{
					RuleID:      "FILE-K8S-003",
					RuleName:    "Security Context",
					Category:    "File Compliance",
					Severity:    "high",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' has no securityContext", containerName),
					Remediation: "Add securityContext with runAsNonRoot: true",
				})
			}

			// Check probes
			if container["livenessProbe"] == nil {
				results = append(results, CheckResult{
					RuleID:      "FILE-K8S-004",
					RuleName:    "Liveness Probe",
					Category:    "File Compliance",
					Severity:    "medium",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' has no livenessProbe", containerName),
					Remediation: "Add livenessProbe",
				})
			}
		}
	}

	return results, nil
}

func (c *FileChecker) checkDockerfile(path string) ([]CheckResult, error) {
	var results []CheckResult
	resource := path

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	hasUser := false
	hasHealthcheck := false
	usesLatest := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		upperLine := strings.ToUpper(line)

		// Check USER directive
		if strings.HasPrefix(upperLine, "USER ") {
			hasUser = true
		}

		// Check HEALTHCHECK
		if strings.HasPrefix(upperLine, "HEALTHCHECK ") {
			hasHealthcheck = true
		}

		// Check FROM with latest
		if strings.HasPrefix(upperLine, "FROM ") {
			if strings.HasSuffix(line, ":latest") || !strings.Contains(line, ":") {
				usesLatest = true
			}
		}

		// Check for ADD when COPY could be used
		if strings.HasPrefix(upperLine, "ADD ") && !strings.Contains(line, "http") && !strings.Contains(line, ".tar") {
			results = append(results, CheckResult{
				RuleID:      "FILE-DOCKER-001",
				RuleName:    "Use COPY Instead of ADD",
				Category:    "File Compliance",
				Severity:    "low",
				Status:      StatusFailed,
				Resource:    resource,
				Message:     "Use COPY instead of ADD for local files",
				Remediation: "Replace ADD with COPY for local files",
			})
		}

		// Check for curl/wget without cleanup
		if strings.Contains(line, "curl") || strings.Contains(line, "wget") {
			if !strings.Contains(line, "&&") || !strings.Contains(line, "rm") {
				results = append(results, CheckResult{
					RuleID:      "FILE-DOCKER-002",
					RuleName:    "Clean Up Downloads",
					Category:    "File Compliance",
					Severity:    "low",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     "Downloaded files should be cleaned up in same layer",
					Remediation: "Combine download and cleanup in single RUN command",
				})
			}
		}
	}

	if !hasUser {
		results = append(results, CheckResult{
			RuleID:      "FILE-DOCKER-003",
			RuleName:    "USER Directive",
			Category:    "File Compliance",
			Severity:    "high",
			Status:      StatusFailed,
			Resource:    resource,
			Message:     "Dockerfile has no USER directive",
			Remediation: "Add USER directive to run as non-root",
		})
	}

	if !hasHealthcheck {
		results = append(results, CheckResult{
			RuleID:      "FILE-DOCKER-004",
			RuleName:    "HEALTHCHECK Directive",
			Category:    "File Compliance",
			Severity:    "medium",
			Status:      StatusFailed,
			Resource:    resource,
			Message:     "Dockerfile has no HEALTHCHECK",
			Remediation: "Add HEALTHCHECK directive",
		})
	}

	if usesLatest {
		results = append(results, CheckResult{
			RuleID:      "FILE-DOCKER-005",
			RuleName:    "Specific Base Image Tag",
			Category:    "File Compliance",
			Severity:    "medium",
			Status:      StatusFailed,
			Resource:    resource,
			Message:     "Base image uses 'latest' or no tag",
			Remediation: "Use specific version tag for base image",
		})
	}

	return results, nil
}

func (c *FileChecker) checkDockerCompose(path string) ([]CheckResult, error) {
	var results []CheckResult
	resource := path

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var compose map[string]interface{}
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, err
	}

	services, _ := compose["services"].(map[string]interface{})
	for serviceName, svc := range services {
		service, _ := svc.(map[string]interface{})

		// Check privileged
		if privileged, ok := service["privileged"].(bool); ok && privileged {
			results = append(results, CheckResult{
				RuleID:      "FILE-COMPOSE-001",
				RuleName:    "No Privileged Services",
				Category:    "File Compliance",
				Severity:    "critical",
				Status:      StatusFailed,
				Resource:    resource,
				Message:     fmt.Sprintf("Service '%s' is privileged", serviceName),
				Remediation: "Remove privileged: true",
			})
		}

		// Check network_mode: host
		if networkMode, ok := service["network_mode"].(string); ok && networkMode == "host" {
			results = append(results, CheckResult{
				RuleID:      "FILE-COMPOSE-002",
				RuleName:    "No Host Network",
				Category:    "File Compliance",
				Severity:    "high",
				Status:      StatusFailed,
				Resource:    resource,
				Message:     fmt.Sprintf("Service '%s' uses host network", serviceName),
				Remediation: "Use bridge network",
			})
		}

		// Check for restart policy
		if service["restart"] == nil && service["deploy"] == nil {
			results = append(results, CheckResult{
				RuleID:      "FILE-COMPOSE-003",
				RuleName:    "Restart Policy",
				Category:    "File Compliance",
				Severity:    "low",
				Status:      StatusFailed,
				Resource:    resource,
				Message:     fmt.Sprintf("Service '%s' has no restart policy", serviceName),
				Remediation: "Add restart: unless-stopped",
			})
		}

		// Check image tag
		if image, ok := service["image"].(string); ok {
			if strings.HasSuffix(image, ":latest") || !strings.Contains(image, ":") {
				results = append(results, CheckResult{
					RuleID:      "FILE-COMPOSE-004",
					RuleName:    "Specific Image Tag",
					Category:    "File Compliance",
					Severity:    "medium",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Service '%s' uses latest or no tag", serviceName),
					Remediation: "Use specific image tag",
				})
			}
		}
	}

	return results, nil
}

func getNestedMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if nested, ok := v.(map[string]interface{}); ok {
			return nested
		}
	}
	return nil
}


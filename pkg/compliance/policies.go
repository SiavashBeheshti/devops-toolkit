package compliance

// GetBuiltinPolicies returns all built-in compliance policies
func GetBuiltinPolicies() []Policy {
	return []Policy{
		// Kubernetes Security
		{
			ID:          "K8S-SEC-001",
			Name:        "No Privileged Containers",
			Category:    "Kubernetes Security",
			Severity:    "critical",
			Description: "Containers should not run in privileged mode as it grants full host access",
			Remediation: "Set securityContext.privileged to false",
		},
		{
			ID:          "K8S-SEC-002",
			Name:        "Run as Non-Root",
			Category:    "Kubernetes Security",
			Severity:    "high",
			Description: "Containers should run as non-root user to limit potential damage",
			Remediation: "Set securityContext.runAsNonRoot to true and specify runAsUser",
		},
		{
			ID:          "K8S-SEC-003",
			Name:        "Read-Only Root Filesystem",
			Category:    "Kubernetes Security",
			Severity:    "medium",
			Description: "Container root filesystem should be read-only to prevent modifications",
			Remediation: "Set securityContext.readOnlyRootFilesystem to true",
		},
		{
			ID:          "K8S-SEC-004",
			Name:        "No Host Network",
			Category:    "Kubernetes Security",
			Severity:    "high",
			Description: "Pods should not use the host network namespace",
			Remediation: "Set hostNetwork to false",
		},
		{
			ID:          "K8S-SEC-005",
			Name:        "No Host PID",
			Category:    "Kubernetes Security",
			Severity:    "high",
			Description: "Pods should not share the host PID namespace",
			Remediation: "Set hostPID to false",
		},

		// Kubernetes Best Practices
		{
			ID:          "K8S-IMG-001",
			Name:        "No Latest Tag",
			Category:    "Kubernetes Best Practices",
			Severity:    "medium",
			Description: "Images should use specific tags instead of 'latest'",
			Remediation: "Use specific version tags for container images",
		},
		{
			ID:          "K8S-PROBE-001",
			Name:        "Liveness Probe",
			Category:    "Kubernetes Best Practices",
			Severity:    "medium",
			Description: "Containers should have liveness probes for automatic restart",
			Remediation: "Add livenessProbe to container spec",
		},
		{
			ID:          "K8S-PROBE-002",
			Name:        "Readiness Probe",
			Category:    "Kubernetes Best Practices",
			Severity:    "medium",
			Description: "Containers should have readiness probes for traffic management",
			Remediation: "Add readinessProbe to container spec",
		},

		// Kubernetes Resources
		{
			ID:          "K8S-RES-001",
			Name:        "CPU Limits",
			Category:    "Kubernetes Resources",
			Severity:    "medium",
			Description: "Containers should have CPU limits to prevent resource starvation",
			Remediation: "Set resources.limits.cpu",
		},
		{
			ID:          "K8S-RES-002",
			Name:        "Memory Limits",
			Category:    "Kubernetes Resources",
			Severity:    "high",
			Description: "Containers should have memory limits to prevent OOM issues",
			Remediation: "Set resources.limits.memory",
		},

		// Kubernetes Network
		{
			ID:          "K8S-NET-001",
			Name:        "Network Policies",
			Category:    "Kubernetes Network",
			Severity:    "medium",
			Description: "Namespaces should have NetworkPolicies to restrict traffic",
			Remediation: "Define NetworkPolicies for the namespace",
		},

		// Kubernetes RBAC
		{
			ID:          "K8S-RBAC-001",
			Name:        "Cluster Admin Bindings",
			Category:    "Kubernetes RBAC",
			Severity:    "high",
			Description: "Avoid granting cluster-admin role to non-system users",
			Remediation: "Use more restrictive roles",
		},

		// Docker Security
		{
			ID:          "DOCKER-SEC-001",
			Name:        "No Privileged Containers",
			Category:    "Docker Security",
			Severity:    "critical",
			Description: "Containers should not run in privileged mode",
			Remediation: "Remove --privileged flag",
		},
		{
			ID:          "DOCKER-SEC-002",
			Name:        "Non-Root User",
			Category:    "Docker Security",
			Severity:    "high",
			Description: "Containers should run as non-root user",
			Remediation: "Use USER directive in Dockerfile or --user flag",
		},
		{
			ID:          "DOCKER-SEC-003",
			Name:        "No Host Network",
			Category:    "Docker Security",
			Severity:    "high",
			Description: "Containers should not use host network",
			Remediation: "Use bridge or custom network",
		},
		{
			ID:          "DOCKER-SEC-004",
			Name:        "No Host PID",
			Category:    "Docker Security",
			Severity:    "high",
			Description: "Containers should not share host PID namespace",
			Remediation: "Remove --pid=host flag",
		},
		{
			ID:          "DOCKER-SEC-005",
			Name:        "No Dangerous Capabilities",
			Category:    "Docker Security",
			Severity:    "high",
			Description: "Containers should not have dangerous Linux capabilities",
			Remediation: "Remove unnecessary --cap-add flags",
		},
		{
			ID:          "DOCKER-SEC-006",
			Name:        "Read-Only Root Filesystem",
			Category:    "Docker Security",
			Severity:    "medium",
			Description: "Container root filesystem should be read-only",
			Remediation: "Use --read-only flag",
		},

		// Docker Resources
		{
			ID:          "DOCKER-RES-001",
			Name:        "Memory Limits",
			Category:    "Docker Resources",
			Severity:    "medium",
			Description: "Containers should have memory limits",
			Remediation: "Set --memory flag",
		},
		{
			ID:          "DOCKER-RES-002",
			Name:        "CPU Limits",
			Category:    "Docker Resources",
			Severity:    "low",
			Description: "Containers should have CPU limits",
			Remediation: "Set --cpus or --cpu-quota flag",
		},

		// Docker Configuration
		{
			ID:          "DOCKER-CFG-001",
			Name:        "Restart Policy",
			Category:    "Docker Configuration",
			Severity:    "low",
			Description: "Containers should have a restart policy",
			Remediation: "Set --restart=unless-stopped",
		},
		{
			ID:          "DOCKER-CFG-002",
			Name:        "Health Check",
			Category:    "Docker Configuration",
			Severity:    "medium",
			Description: "Containers should have health checks",
			Remediation: "Add HEALTHCHECK in Dockerfile or --health-cmd",
		},

		// Docker Images
		{
			ID:          "DOCKER-IMG-001",
			Name:        "No Latest Tag",
			Category:    "Docker Images",
			Severity:    "medium",
			Description: "Images should use specific tags",
			Remediation: "Use specific version tags",
		},
		{
			ID:          "DOCKER-IMG-002",
			Name:        "Image Size",
			Category:    "Docker Images",
			Severity:    "low",
			Description: "Images should not be excessively large",
			Remediation: "Use multi-stage builds or smaller base images",
		},
		{
			ID:          "DOCKER-IMG-003",
			Name:        "Non-Root User in Image",
			Category:    "Docker Images",
			Severity:    "medium",
			Description: "Images should define a non-root user",
			Remediation: "Add USER directive in Dockerfile",
		},

		// File Compliance
		{
			ID:          "FILE-K8S-001",
			Name:        "No Latest Tag in Manifests",
			Category:    "File Compliance",
			Severity:    "medium",
			Description: "Kubernetes manifests should use specific image tags",
			Remediation: "Use specific version tags",
		},
		{
			ID:          "FILE-K8S-002",
			Name:        "Resource Limits in Manifests",
			Category:    "File Compliance",
			Severity:    "medium",
			Description: "Kubernetes manifests should define resource limits",
			Remediation: "Add resources.limits",
		},
		{
			ID:          "FILE-K8S-003",
			Name:        "Security Context in Manifests",
			Category:    "File Compliance",
			Severity:    "high",
			Description: "Kubernetes manifests should define security context",
			Remediation: "Add securityContext",
		},
		{
			ID:          "FILE-DOCKER-003",
			Name:        "USER in Dockerfile",
			Category:    "File Compliance",
			Severity:    "high",
			Description: "Dockerfiles should define a non-root USER",
			Remediation: "Add USER directive",
		},
		{
			ID:          "FILE-DOCKER-004",
			Name:        "HEALTHCHECK in Dockerfile",
			Category:    "File Compliance",
			Severity:    "medium",
			Description: "Dockerfiles should define a HEALTHCHECK",
			Remediation: "Add HEALTHCHECK directive",
		},
		{
			ID:          "FILE-COMPOSE-001",
			Name:        "No Privileged in Compose",
			Category:    "File Compliance",
			Severity:    "critical",
			Description: "Docker Compose services should not be privileged",
			Remediation: "Remove privileged: true",
		},
	}
}


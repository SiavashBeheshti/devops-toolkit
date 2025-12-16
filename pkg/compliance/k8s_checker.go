package compliance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sChecker checks Kubernetes resources for compliance
type K8sChecker struct {
	opts      CheckOptions
	clientset *kubernetes.Clientset
}

// NewK8sChecker creates a new Kubernetes checker
func NewK8sChecker(opts CheckOptions) *K8sChecker {
	return &K8sChecker{opts: opts}
}

// Run runs the Kubernetes compliance checks
func (c *K8sChecker) Run(ctx context.Context) ([]CheckResult, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	var results []CheckResult

	// Pod security checks
	podResults, err := c.checkPodSecurity(ctx)
	if err == nil {
		results = append(results, podResults...)
	}

	// Container checks
	containerResults, err := c.checkContainers(ctx)
	if err == nil {
		results = append(results, containerResults...)
	}

	// Resource limit checks
	resourceResults, err := c.checkResourceLimits(ctx)
	if err == nil {
		results = append(results, resourceResults...)
	}

	// Network policy checks
	networkResults, err := c.checkNetworkPolicies(ctx)
	if err == nil {
		results = append(results, networkResults...)
	}

	// RBAC checks
	rbacResults, err := c.checkRBAC(ctx)
	if err == nil {
		results = append(results, rbacResults...)
	}

	return c.filterResults(results), nil
}

func (c *K8sChecker) initClient() error {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	c.clientset = clientset
	return nil
}

func (c *K8sChecker) checkPodSecurity(ctx context.Context) ([]CheckResult, error) {
	var results []CheckResult

	pods, err := c.clientset.CoreV1().Pods(c.opts.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		resource := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

		// Check for privileged containers
		for _, container := range pod.Spec.Containers {
			if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
				results = append(results, CheckResult{
					RuleID:      "K8S-SEC-001",
					RuleName:    "No Privileged Containers",
					Category:    "Kubernetes Security",
					Severity:    "critical",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' is running in privileged mode", container.Name),
					Remediation: "Set securityContext.privileged to false",
				})
			} else {
				results = append(results, CheckResult{
					RuleID:   "K8S-SEC-001",
					RuleName: "No Privileged Containers",
					Category: "Kubernetes Security",
					Severity: "critical",
					Status:   StatusPassed,
					Resource: resource,
					Message:  fmt.Sprintf("Container '%s' is not privileged", container.Name),
				})
			}

			// Check for root user
			if container.SecurityContext == nil || container.SecurityContext.RunAsNonRoot == nil || !*container.SecurityContext.RunAsNonRoot {
				results = append(results, CheckResult{
					RuleID:      "K8S-SEC-002",
					RuleName:    "Run as Non-Root",
					Category:    "Kubernetes Security",
					Severity:    "high",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' may run as root", container.Name),
					Remediation: "Set securityContext.runAsNonRoot to true",
				})
			}

			// Check for read-only root filesystem
			if container.SecurityContext == nil || container.SecurityContext.ReadOnlyRootFilesystem == nil || !*container.SecurityContext.ReadOnlyRootFilesystem {
				results = append(results, CheckResult{
					RuleID:      "K8S-SEC-003",
					RuleName:    "Read-Only Root Filesystem",
					Category:    "Kubernetes Security",
					Severity:    "medium",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' has writable root filesystem", container.Name),
					Remediation: "Set securityContext.readOnlyRootFilesystem to true",
				})
			}
		}

		// Check host network
		if pod.Spec.HostNetwork {
			results = append(results, CheckResult{
				RuleID:      "K8S-SEC-004",
				RuleName:    "No Host Network",
				Category:    "Kubernetes Security",
				Severity:    "high",
				Status:      StatusFailed,
				Resource:    resource,
				Message:     "Pod is using host network",
				Remediation: "Set hostNetwork to false",
			})
		}

		// Check host PID
		if pod.Spec.HostPID {
			results = append(results, CheckResult{
				RuleID:      "K8S-SEC-005",
				RuleName:    "No Host PID",
				Category:    "Kubernetes Security",
				Severity:    "high",
				Status:      StatusFailed,
				Resource:    resource,
				Message:     "Pod is using host PID namespace",
				Remediation: "Set hostPID to false",
			})
		}
	}

	return results, nil
}

func (c *K8sChecker) checkContainers(ctx context.Context) ([]CheckResult, error) {
	var results []CheckResult

	pods, err := c.clientset.CoreV1().Pods(c.opts.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		resource := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

		for _, container := range pod.Spec.Containers {
			// Check for latest tag
			if strings.HasSuffix(container.Image, ":latest") || !strings.Contains(container.Image, ":") {
				results = append(results, CheckResult{
					RuleID:      "K8S-IMG-001",
					RuleName:    "No Latest Tag",
					Category:    "Kubernetes Best Practices",
					Severity:    "medium",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' uses latest or no tag: %s", container.Name, container.Image),
					Remediation: "Use specific image tags",
				})
			}

			// Check image pull policy
			if container.ImagePullPolicy == corev1.PullAlways {
				results = append(results, CheckResult{
					RuleID:   "K8S-IMG-002",
					RuleName: "Image Pull Policy",
					Category: "Kubernetes Best Practices",
					Severity: "low",
					Status:   StatusPassed,
					Resource: resource,
					Message:  fmt.Sprintf("Container '%s' has ImagePullPolicy: Always", container.Name),
				})
			}

			// Check for liveness probe
			if container.LivenessProbe == nil {
				results = append(results, CheckResult{
					RuleID:      "K8S-PROBE-001",
					RuleName:    "Liveness Probe",
					Category:    "Kubernetes Best Practices",
					Severity:    "medium",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' has no liveness probe", container.Name),
					Remediation: "Add a livenessProbe to the container",
				})
			}

			// Check for readiness probe
			if container.ReadinessProbe == nil {
				results = append(results, CheckResult{
					RuleID:      "K8S-PROBE-002",
					RuleName:    "Readiness Probe",
					Category:    "Kubernetes Best Practices",
					Severity:    "medium",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' has no readiness probe", container.Name),
					Remediation: "Add a readinessProbe to the container",
				})
			}
		}
	}

	return results, nil
}

func (c *K8sChecker) checkResourceLimits(ctx context.Context) ([]CheckResult, error) {
	var results []CheckResult

	pods, err := c.clientset.CoreV1().Pods(c.opts.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		resource := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

		for _, container := range pod.Spec.Containers {
			// Check CPU limits
			if container.Resources.Limits.Cpu().IsZero() {
				results = append(results, CheckResult{
					RuleID:      "K8S-RES-001",
					RuleName:    "CPU Limits",
					Category:    "Kubernetes Resources",
					Severity:    "medium",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' has no CPU limit", container.Name),
					Remediation: "Set resources.limits.cpu",
				})
			}

			// Check memory limits
			if container.Resources.Limits.Memory().IsZero() {
				results = append(results, CheckResult{
					RuleID:      "K8S-RES-002",
					RuleName:    "Memory Limits",
					Category:    "Kubernetes Resources",
					Severity:    "high",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' has no memory limit", container.Name),
					Remediation: "Set resources.limits.memory",
				})
			}

			// Check CPU requests
			if container.Resources.Requests.Cpu().IsZero() {
				results = append(results, CheckResult{
					RuleID:      "K8S-RES-003",
					RuleName:    "CPU Requests",
					Category:    "Kubernetes Resources",
					Severity:    "low",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' has no CPU request", container.Name),
					Remediation: "Set resources.requests.cpu",
				})
			}

			// Check memory requests
			if container.Resources.Requests.Memory().IsZero() {
				results = append(results, CheckResult{
					RuleID:      "K8S-RES-004",
					RuleName:    "Memory Requests",
					Category:    "Kubernetes Resources",
					Severity:    "low",
					Status:      StatusFailed,
					Resource:    resource,
					Message:     fmt.Sprintf("Container '%s' has no memory request", container.Name),
					Remediation: "Set resources.requests.memory",
				})
			}
		}
	}

	return results, nil
}

func (c *K8sChecker) checkNetworkPolicies(ctx context.Context) ([]CheckResult, error) {
	var results []CheckResult

	// Get all namespaces
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, ns := range namespaces.Items {
		// Skip system namespaces
		if strings.HasPrefix(ns.Name, "kube-") {
			continue
		}

		if c.opts.Namespace != "" && ns.Name != c.opts.Namespace {
			continue
		}

		// Check if namespace has network policies
		policies, err := c.clientset.NetworkingV1().NetworkPolicies(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}

		if len(policies.Items) == 0 {
			results = append(results, CheckResult{
				RuleID:      "K8S-NET-001",
				RuleName:    "Network Policies",
				Category:    "Kubernetes Network",
				Severity:    "medium",
				Status:      StatusFailed,
				Resource:    ns.Name,
				Message:     fmt.Sprintf("Namespace '%s' has no NetworkPolicies", ns.Name),
				Remediation: "Define NetworkPolicies to restrict pod traffic",
			})
		} else {
			results = append(results, CheckResult{
				RuleID:   "K8S-NET-001",
				RuleName: "Network Policies",
				Category: "Kubernetes Network",
				Severity: "medium",
				Status:   StatusPassed,
				Resource: ns.Name,
				Message:  fmt.Sprintf("Namespace '%s' has %d NetworkPolicies", ns.Name, len(policies.Items)),
			})
		}
	}

	return results, nil
}

func (c *K8sChecker) checkRBAC(ctx context.Context) ([]CheckResult, error) {
	var results []CheckResult

	// Check for cluster-admin bindings
	bindings, err := c.clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, binding := range bindings.Items {
		if binding.RoleRef.Name == "cluster-admin" {
			// Skip system bindings
			if strings.HasPrefix(binding.Name, "system:") {
				continue
			}

			results = append(results, CheckResult{
				RuleID:      "K8S-RBAC-001",
				RuleName:    "Cluster Admin Bindings",
				Category:    "Kubernetes RBAC",
				Severity:    "high",
				Status:      StatusFailed,
				Resource:    binding.Name,
				Message:     fmt.Sprintf("ClusterRoleBinding '%s' grants cluster-admin", binding.Name),
				Remediation: "Use more restrictive roles",
			})
		}
	}

	return results, nil
}

func (c *K8sChecker) filterResults(results []CheckResult) []CheckResult {
	if len(c.opts.SkipRules) == 0 && len(c.opts.OnlyRules) == 0 && c.opts.MinSeverity == "" {
		return results
	}

	var filtered []CheckResult
	for _, r := range results {
		// Skip rules
		skip := false
		for _, skipRule := range c.opts.SkipRules {
			if r.RuleID == skipRule {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		// Only rules
		if len(c.opts.OnlyRules) > 0 {
			found := false
			for _, onlyRule := range c.opts.OnlyRules {
				if r.RuleID == onlyRule {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Min severity
		if c.opts.MinSeverity != "" && !meetsMinSeverity(r.Severity, c.opts.MinSeverity) {
			continue
		}

		filtered = append(filtered, r)
	}

	return filtered
}

func meetsMinSeverity(severity, minSeverity string) bool {
	levels := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}

	return levels[severity] >= levels[minSeverity]
}


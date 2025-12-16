package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes clientset
type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfigPath, context string) (*Client, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		if kubeconfigPath == "" {
			kubeconfigPath = os.Getenv("KUBECONFIG")
			if kubeconfigPath == "" {
				home, _ := os.UserHomeDir()
				kubeconfigPath = filepath.Join(home, ".kube", "config")
			}
		}

		loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
		configOverrides := &clientcmd.ConfigOverrides{}
		if context != "" {
			configOverrides.CurrentContext = context
		}

		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		config, err = kubeConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}

// ClusterInfo contains cluster information
type ClusterInfo struct {
	Name       string
	Server     string
	K8sVersion string
}

// GetClusterInfo returns cluster information
func (c *Client) GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	version, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}

	return &ClusterInfo{
		Name:       c.config.Host,
		Server:     c.config.Host,
		K8sVersion: version.GitVersion,
	}, nil
}

// NodeHealth contains node health information
type NodeHealth struct {
	Total   int
	Ready   int
	Healthy bool
}

// GetNodeHealth returns node health status
func (c *Client) GetNodeHealth(ctx context.Context) (*NodeHealth, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	health := &NodeHealth{
		Total: len(nodes.Items),
	}

	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				health.Ready++
				break
			}
		}
	}

	health.Healthy = health.Ready == health.Total
	return health, nil
}

// PodHealth contains pod health information
type PodHealth struct {
	Running int
	Pending int
	Failed  int
	Total   int
}

// GetPodHealth returns pod health status
func (c *Client) GetPodHealth(ctx context.Context, namespace string) (*PodHealth, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	health := &PodHealth{
		Total: len(pods.Items),
	}

	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case corev1.PodRunning:
			health.Running++
		case corev1.PodPending:
			health.Pending++
		case corev1.PodFailed:
			health.Failed++
		}
	}

	return health, nil
}

// PVCHealth contains PVC health information
type PVCHealth struct {
	Bound   int
	Pending int
	Total   int
}

// GetPVCHealth returns PVC health status
func (c *Client) GetPVCHealth(ctx context.Context, namespace string) (*PVCHealth, error) {
	pvcs, err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	health := &PVCHealth{
		Total: len(pvcs.Items),
	}

	for _, pvc := range pvcs.Items {
		switch pvc.Status.Phase {
		case corev1.ClaimBound:
			health.Bound++
		case corev1.ClaimPending:
			health.Pending++
		}
	}

	return health, nil
}

// DeploymentHealth contains deployment health information
type DeploymentHealth struct {
	Total       int
	Ready       int
	Unavailable int
}

// GetDeploymentHealth returns deployment health status
func (c *Client) GetDeploymentHealth(ctx context.Context, namespace string) (*DeploymentHealth, error) {
	deployments, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	health := &DeploymentHealth{
		Total: len(deployments.Items),
	}

	for _, dep := range deployments.Items {
		if dep.Status.ReadyReplicas == *dep.Spec.Replicas {
			health.Ready++
		}
		health.Unavailable += int(dep.Status.UnavailableReplicas)
	}

	return health, nil
}

// ServiceHealth contains service health information
type ServiceHealth struct {
	ClusterIP    int
	LoadBalancer int
	NodePort     int
	Total        int
}

// GetServiceHealth returns service health status
func (c *Client) GetServiceHealth(ctx context.Context, namespace string) (*ServiceHealth, error) {
	services, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	health := &ServiceHealth{
		Total: len(services.Items),
	}

	for _, svc := range services.Items {
		switch svc.Spec.Type {
		case corev1.ServiceTypeClusterIP:
			health.ClusterIP++
		case corev1.ServiceTypeLoadBalancer:
			health.LoadBalancer++
		case corev1.ServiceTypeNodePort:
			health.NodePort++
		}
	}

	return health, nil
}

// ResourceUtilization contains resource utilization information
type ResourceUtilization struct {
	CPUUsed        int64
	CPUCapacity    int64
	MemoryUsed     int64
	MemoryCapacity int64
}

// GetResourceUtilization returns resource utilization
func (c *Client) GetResourceUtilization(ctx context.Context) (*ResourceUtilization, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	util := &ResourceUtilization{}

	for _, node := range nodes.Items {
		util.CPUCapacity += node.Status.Capacity.Cpu().MilliValue()
		util.MemoryCapacity += node.Status.Capacity.Memory().Value()
	}

	// Get pod resource requests as a proxy for usage
	pods, err := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: "status.phase=Running",
	})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			util.CPUUsed += container.Resources.Requests.Cpu().MilliValue()
			util.MemoryUsed += container.Resources.Requests.Memory().Value()
		}
	}

	return util, nil
}

// EventInfo contains event information
type EventInfo struct {
	Type          string
	Reason        string
	Object        string
	Kind          string
	Message       string
	Count         int32
	LastTimestamp time.Time
}

// GetWarningEvents returns recent warning events
func (c *Client) GetWarningEvents(ctx context.Context, namespace string, limit int) ([]EventInfo, error) {
	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: "type=Warning",
	})
	if err != nil {
		return nil, err
	}

	// Sort by last timestamp descending
	sort.Slice(events.Items, func(i, j int) bool {
		return events.Items[i].LastTimestamp.After(events.Items[j].LastTimestamp.Time)
	})

	var result []EventInfo
	for i, event := range events.Items {
		if i >= limit {
			break
		}
		// Only include recent events (last hour)
		if time.Since(event.LastTimestamp.Time) > time.Hour {
			continue
		}
		result = append(result, EventInfo{
			Type:          event.Type,
			Reason:        event.Reason,
			Object:        event.InvolvedObject.Name,
			Kind:          event.InvolvedObject.Kind,
			Message:       event.Message,
			Count:         event.Count,
			LastTimestamp: event.LastTimestamp.Time,
		})
	}

	return result, nil
}

// PodInfo contains pod information
type PodInfo struct {
	Name            string
	Namespace       string
	Status          string
	ReadyContainers int
	TotalContainers int
	Restarts        int32
	Node            string
	IP              string
	CreationTime    time.Time
}

// ListPods lists pods with enhanced information
func (c *Client) ListPods(ctx context.Context, namespace, labelSelector string) ([]PodInfo, error) {
	opts := metav1.ListOptions{}
	if labelSelector != "" {
		opts.LabelSelector = labelSelector
	}

	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	var result []PodInfo
	for _, pod := range pods.Items {
		info := PodInfo{
			Name:            pod.Name,
			Namespace:       pod.Namespace,
			TotalContainers: len(pod.Spec.Containers),
			Node:            pod.Spec.NodeName,
			IP:              pod.Status.PodIP,
			CreationTime:    pod.CreationTimestamp.Time,
		}

		// Calculate ready containers and restarts
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Ready {
				info.ReadyContainers++
			}
			info.Restarts += cs.RestartCount
		}

		// Determine status
		info.Status = string(pod.Status.Phase)
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
				info.Status = cs.State.Waiting.Reason
				break
			}
			if cs.State.Terminated != nil && cs.State.Terminated.Reason != "" {
				info.Status = cs.State.Terminated.Reason
				break
			}
		}

		// Check for eviction
		if pod.Status.Reason == "Evicted" {
			info.Status = "Evicted"
		}

		result = append(result, info)
	}

	return result, nil
}

// NodeInfo contains node information
type NodeInfo struct {
	Name               string
	Ready              bool
	Roles              string
	KubeletVersion     string
	InternalIP         string
	ExternalIP         string
	OSImage            string
	KernelVersion      string
	ContainerRuntime   string
	CPUCapacity        int64
	MemoryCapacity     int64
	CPUUsagePercent    float64
	MemoryUsagePercent float64
	MemoryPressure     bool
	DiskPressure       bool
	PIDPressure        bool
	CreationTime       time.Time
}

// ListNodes lists cluster nodes with enhanced information
func (c *Client) ListNodes(ctx context.Context) ([]NodeInfo, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []NodeInfo
	for _, node := range nodes.Items {
		info := NodeInfo{
			Name:             node.Name,
			KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
			OSImage:          node.Status.NodeInfo.OSImage,
			KernelVersion:    node.Status.NodeInfo.KernelVersion,
			ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
			CPUCapacity:      node.Status.Capacity.Cpu().MilliValue(),
			MemoryCapacity:   node.Status.Capacity.Memory().Value(),
			CreationTime:     node.CreationTimestamp.Time,
		}

		// Get roles
		var roles []string
		for label := range node.Labels {
			if strings.HasPrefix(label, "node-role.kubernetes.io/") {
				role := strings.TrimPrefix(label, "node-role.kubernetes.io/")
				roles = append(roles, role)
			}
		}
		if len(roles) == 0 {
			roles = append(roles, "<none>")
		}
		info.Roles = strings.Join(roles, ",")

		// Get addresses
		for _, addr := range node.Status.Addresses {
			switch addr.Type {
			case corev1.NodeInternalIP:
				info.InternalIP = addr.Address
			case corev1.NodeExternalIP:
				info.ExternalIP = addr.Address
			}
		}

		// Get conditions
		for _, condition := range node.Status.Conditions {
			switch condition.Type {
			case corev1.NodeReady:
				info.Ready = condition.Status == corev1.ConditionTrue
			case corev1.NodeMemoryPressure:
				info.MemoryPressure = condition.Status == corev1.ConditionTrue
			case corev1.NodeDiskPressure:
				info.DiskPressure = condition.Status == corev1.ConditionTrue
			case corev1.NodePIDPressure:
				info.PIDPressure = condition.Status == corev1.ConditionTrue
			}
		}

		result = append(result, info)
	}

	return result, nil
}

// FindCompletedPods finds completed pods
func (c *Client) FindCompletedPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	pods, err := c.ListPods(ctx, namespace, "")
	if err != nil {
		return nil, err
	}

	var result []PodInfo
	for _, pod := range pods {
		if pod.Status == "Succeeded" || pod.Status == "Completed" {
			result = append(result, pod)
		}
	}
	return result, nil
}

// FindFailedPods finds failed pods
func (c *Client) FindFailedPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	pods, err := c.ListPods(ctx, namespace, "")
	if err != nil {
		return nil, err
	}

	var result []PodInfo
	for _, pod := range pods {
		status := strings.ToLower(pod.Status)
		if strings.Contains(status, "error") || strings.Contains(status, "failed") ||
			strings.Contains(status, "crash") || strings.Contains(status, "backoff") {
			result = append(result, pod)
		}
	}
	return result, nil
}

// FindEvictedPods finds evicted pods
func (c *Client) FindEvictedPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	pods, err := c.ListPods(ctx, namespace, "")
	if err != nil {
		return nil, err
	}

	var result []PodInfo
	for _, pod := range pods {
		if pod.Status == "Evicted" {
			result = append(result, pod)
		}
	}
	return result, nil
}

// DeletePods deletes the specified pods
func (c *Client) DeletePods(ctx context.Context, pods []PodInfo) (int, error) {
	deleted := 0
	for _, pod := range pods {
		err := c.clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if err == nil {
			deleted++
		}
	}
	return deleted, nil
}

// JobInfo contains job information
type JobInfo struct {
	Name      string
	Namespace string
}

// FindCompletedJobs finds completed jobs
func (c *Client) FindCompletedJobs(ctx context.Context, namespace string) ([]JobInfo, error) {
	jobs, err := c.clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []JobInfo
	for _, job := range jobs.Items {
		if job.Status.Succeeded > 0 && job.Status.Active == 0 {
			result = append(result, JobInfo{
				Name:      job.Name,
				Namespace: job.Namespace,
			})
		}
	}
	return result, nil
}

// DeleteJobs deletes the specified jobs
func (c *Client) DeleteJobs(ctx context.Context, jobs []JobInfo) (int, error) {
	deleted := 0
	propagation := metav1.DeletePropagationBackground
	for _, job := range jobs {
		err := c.clientset.BatchV1().Jobs(job.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{
			PropagationPolicy: &propagation,
		})
		if err == nil {
			deleted++
		}
	}
	return deleted, nil
}

// ReplicaSetInfo contains ReplicaSet information
type ReplicaSetInfo struct {
	Name      string
	Namespace string
}

// FindOrphanedReplicaSets finds orphaned ReplicaSets
func (c *Client) FindOrphanedReplicaSets(ctx context.Context, namespace string) ([]ReplicaSetInfo, error) {
	replicaSets, err := c.clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []ReplicaSetInfo
	for _, rs := range replicaSets.Items {
		// Orphaned RS have 0 replicas and no owner
		if rs.Status.Replicas == 0 && len(rs.OwnerReferences) == 0 {
			result = append(result, ReplicaSetInfo{
				Name:      rs.Name,
				Namespace: rs.Namespace,
			})
		}
	}
	return result, nil
}

// DeleteReplicaSets deletes the specified ReplicaSets
func (c *Client) DeleteReplicaSets(ctx context.Context, replicaSets []ReplicaSetInfo) (int, error) {
	deleted := 0
	for _, rs := range replicaSets {
		err := c.clientset.AppsV1().ReplicaSets(rs.Namespace).Delete(ctx, rs.Name, metav1.DeleteOptions{})
		if err == nil {
			deleted++
		}
	}
	return deleted, nil
}

// EventFilter contains event filter options
type EventFilter struct {
	Type   string
	Reason string
	Object string
	Limit  int
}

// ListEvents lists events with filters
func (c *Client) ListEvents(ctx context.Context, namespace string, filter EventFilter) ([]EventInfo, error) {
	opts := metav1.ListOptions{}
	if filter.Type != "" {
		opts.FieldSelector = "type=" + filter.Type
	}

	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Sort by last timestamp descending
	sort.Slice(events.Items, func(i, j int) bool {
		return events.Items[i].LastTimestamp.After(events.Items[j].LastTimestamp.Time)
	})

	var result []EventInfo
	for i, event := range events.Items {
		if filter.Limit > 0 && i >= filter.Limit {
			break
		}

		// Apply filters
		if filter.Reason != "" && !strings.Contains(strings.ToLower(event.Reason), strings.ToLower(filter.Reason)) {
			continue
		}
		if filter.Object != "" && !strings.Contains(strings.ToLower(event.InvolvedObject.Name), strings.ToLower(filter.Object)) {
			continue
		}

		result = append(result, EventInfo{
			Type:          event.Type,
			Reason:        event.Reason,
			Object:        event.InvolvedObject.Name,
			Kind:          event.InvolvedObject.Kind,
			Message:       event.Message,
			Count:         event.Count,
			LastTimestamp: event.LastTimestamp.Time,
		})
	}

	return result, nil
}

// ClusterResources contains cluster resource information
type ClusterResources struct {
	CPURequests        int64
	CPULimits          int64
	CPUAllocatable     int64
	MemoryRequests     int64
	MemoryLimits       int64
	MemoryAllocatable  int64
	PodCount           int
	PodCapacity        int
}

// GetClusterResources returns cluster resource information
func (c *Client) GetClusterResources(ctx context.Context) (*ClusterResources, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	res := &ClusterResources{}

	for _, node := range nodes.Items {
		res.CPUAllocatable += node.Status.Allocatable.Cpu().MilliValue()
		res.MemoryAllocatable += node.Status.Allocatable.Memory().Value()
		res.PodCapacity += int(node.Status.Allocatable.Pods().Value())
	}

	pods, err := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	res.PodCount = len(pods.Items)

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			res.CPURequests += container.Resources.Requests.Cpu().MilliValue()
			res.CPULimits += container.Resources.Limits.Cpu().MilliValue()
			res.MemoryRequests += container.Resources.Requests.Memory().Value()
			res.MemoryLimits += container.Resources.Limits.Memory().Value()
		}
	}

	return res, nil
}

// NamespaceResources contains namespace resource information
type NamespaceResources struct {
	Namespace      string
	PodCount       int
	CPURequests    int64
	MemoryRequests int64
}

// GetNamespaceResources returns resource usage by namespace
func (c *Client) GetNamespaceResources(ctx context.Context) ([]NamespaceResources, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result []NamespaceResources

	for _, ns := range namespaces.Items {
		pods, err := c.clientset.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}

		nsRes := NamespaceResources{
			Namespace: ns.Name,
			PodCount:  len(pods.Items),
		}

		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				nsRes.CPURequests += container.Resources.Requests.Cpu().MilliValue()
				nsRes.MemoryRequests += container.Resources.Requests.Memory().Value()
			}
		}

		if nsRes.PodCount > 0 {
			result = append(result, nsRes)
		}
	}

	// Sort by CPU requests descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].CPURequests > result[j].CPURequests
	})

	return result, nil
}

// TopPods contains top resource consuming pods
type TopPods struct {
	ByCPU    []PodResourceUsage
	ByMemory []PodResourceUsage
}

// PodResourceUsage contains pod resource usage
type PodResourceUsage struct {
	Name          string
	Namespace     string
	CPUUsage      int64
	CPURequest    int64
	MemoryUsage   int64
	MemoryRequest int64
}

// GetTopPods returns top resource consuming pods
func (c *Client) GetTopPods(ctx context.Context, namespace string, limit int) (*TopPods, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: "status.phase=Running",
	})
	if err != nil {
		return nil, err
	}

	var usage []PodResourceUsage

	for _, pod := range pods.Items {
		pu := PodResourceUsage{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		}

		for _, container := range pod.Spec.Containers {
			pu.CPURequest += container.Resources.Requests.Cpu().MilliValue()
			pu.MemoryRequest += container.Resources.Requests.Memory().Value()
			// Use requests as proxy for usage since we don't have metrics-server integration
			pu.CPUUsage += container.Resources.Requests.Cpu().MilliValue()
			pu.MemoryUsage += container.Resources.Requests.Memory().Value()
		}

		usage = append(usage, pu)
	}

	result := &TopPods{}

	// Sort by CPU
	sort.Slice(usage, func(i, j int) bool {
		return usage[i].CPUUsage > usage[j].CPUUsage
	})
	for i := 0; i < limit && i < len(usage); i++ {
		result.ByCPU = append(result.ByCPU, usage[i])
	}

	// Sort by Memory
	sort.Slice(usage, func(i, j int) bool {
		return usage[i].MemoryUsage > usage[j].MemoryUsage
	})
	for i := 0; i < limit && i < len(usage); i++ {
		result.ByMemory = append(result.ByMemory, usage[i])
	}

	return result, nil
}


package completion

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// getK8sClient creates a Kubernetes client for completion
func getK8sClient() (*kubernetes.Clientset, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home, _ := os.UserHomeDir()
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

// NamespaceCompletion provides namespace completion
func NamespaceCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client, err := getK8sClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	namespaces, err := client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, ns := range namespaces.Items {
		if strings.HasPrefix(ns.Name, toComplete) {
			completions = append(completions, ns.Name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// PodCompletion provides pod name completion
func PodCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client, err := getK8sClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get namespace from flag or use all namespaces
	namespace := ""
	if ns := cmd.Flag("namespace"); ns != nil && ns.Value.String() != "" {
		namespace = ns.Value.String()
	}

	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, pod := range pods.Items {
		name := pod.Name
		if namespace == "" {
			// Include namespace prefix when listing all namespaces
			name = pod.Namespace + "/" + pod.Name
		}
		if strings.HasPrefix(name, toComplete) || strings.HasPrefix(pod.Name, toComplete) {
			completions = append(completions, name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// ContainerInPodCompletion provides container name completion for a pod
func ContainerInPodCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		// No pod specified yet
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := getK8sClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	namespace := ""
	if ns := cmd.Flag("namespace"); ns != nil && ns.Value.String() != "" {
		namespace = ns.Value.String()
	}

	// Handle namespace/pod format
	podName := args[0]
	if strings.Contains(podName, "/") {
		parts := strings.SplitN(podName, "/", 2)
		namespace = parts[0]
		podName = parts[1]
	}

	if namespace == "" {
		namespace = "default"
	}

	pod, err := client.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, container := range pod.Spec.Containers {
		if strings.HasPrefix(container.Name, toComplete) {
			completions = append(completions, container.Name)
		}
	}
	for _, container := range pod.Spec.InitContainers {
		if strings.HasPrefix(container.Name, toComplete) {
			completions = append(completions, container.Name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// NodeCompletion provides node name completion
func NodeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client, err := getK8sClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	nodes, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, node := range nodes.Items {
		if strings.HasPrefix(node.Name, toComplete) {
			completions = append(completions, node.Name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// DeploymentCompletion provides deployment name completion
func DeploymentCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client, err := getK8sClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	namespace := ""
	if ns := cmd.Flag("namespace"); ns != nil && ns.Value.String() != "" {
		namespace = ns.Value.String()
	}

	deployments, err := client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, dep := range deployments.Items {
		name := dep.Name
		if namespace == "" {
			name = dep.Namespace + "/" + dep.Name
		}
		if strings.HasPrefix(name, toComplete) || strings.HasPrefix(dep.Name, toComplete) {
			completions = append(completions, name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// ServiceCompletion provides service name completion
func ServiceCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client, err := getK8sClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	namespace := ""
	if ns := cmd.Flag("namespace"); ns != nil && ns.Value.String() != "" {
		namespace = ns.Value.String()
	}

	services, err := client.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, svc := range services.Items {
		name := svc.Name
		if namespace == "" {
			name = svc.Namespace + "/" + svc.Name
		}
		if strings.HasPrefix(name, toComplete) || strings.HasPrefix(svc.Name, toComplete) {
			completions = append(completions, name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// ContextCompletion provides kubernetes context completion
func ContextCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		home, _ := os.UserHomeDir()
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for name := range config.Contexts {
		if strings.HasPrefix(name, toComplete) {
			completions = append(completions, name)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// ResourceTypeCompletion provides completion for Kubernetes resource types
func ResourceTypeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resourceTypes := []string{
		"pods", "pod", "po",
		"deployments", "deployment", "deploy",
		"services", "service", "svc",
		"configmaps", "configmap", "cm",
		"secrets", "secret",
		"namespaces", "namespace", "ns",
		"nodes", "node", "no",
		"persistentvolumeclaims", "persistentvolumeclaim", "pvc",
		"persistentvolumes", "persistentvolume", "pv",
		"replicasets", "replicaset", "rs",
		"statefulsets", "statefulset", "sts",
		"daemonsets", "daemonset", "ds",
		"jobs", "job",
		"cronjobs", "cronjob", "cj",
		"ingresses", "ingress", "ing",
		"events", "event", "ev",
	}

	var completions []string
	for _, rt := range resourceTypes {
		if strings.HasPrefix(rt, toComplete) {
			completions = append(completions, rt)
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}


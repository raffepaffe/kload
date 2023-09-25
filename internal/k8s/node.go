package k8s

import (
	"context"
	"fmt"

	"github.com/raffepaffe/kload/internal/cmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Node struct {
	Name        string
	CPU         int64
	CPULimit    int64
	Memory      int64
	MemoryLimit int64
}

func createNodes(command *cmd.Command) ([]*Node, error) {
	config, err := clientcmd.BuildConfigFromFlags("", command.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("BuildConfigFromFlags error: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("NewForConfig error: %w", err)
	}

	nodes, err := createNodesForClient(clientset, config)
	if err != nil {
		return nil, fmt.Errorf("createNodesForClient error: %w", err)
	}

	return nodes, nil
}

func createNodesForClient(clientset *kubernetes.Clientset, config *rest.Config) ([]*Node, error) {
	n, err := createNodeResources(clientset)
	if err != nil {
		return nil, fmt.Errorf("createNodeResources error: %w", err)
	}

	mc, err := metrics.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("createNewMetrics error: %w", err)
	}
	nodes, err := createNodeMetrics(mc, n)
	if err != nil {
		return nil, fmt.Errorf("createNodesMetric error: %w", err)
	}

	return nodes, nil
}

func createNodeResources(clientset *kubernetes.Clientset) ([]*Node, error) {
	result := make([]*Node, 0)
	nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return result, fmt.Errorf("nodex list error: %w", err)
	}

	for _, node := range nodes.Items {
		n := &Node{
			Name:        node.Name,
			CPULimit:    node.Status.Allocatable.Cpu().MilliValue(),
			MemoryLimit: toMegaBytes(node.Status.Allocatable.Memory().Value()),
		}
		result = append(result, n)
	}

	return result, nil
}

func createNodeMetrics(mc *metrics.Clientset, nodeResources []*Node) ([]*Node, error) {
	nodeMetrics, err := mc.MetricsV1beta1().NodeMetricses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("node metrics error: %w", err)
	}

	for _, node := range nodeResources {
		for _, nodeMetric := range nodeMetrics.Items {
			if node.Name == nodeMetric.Name {
				node.CPU = nodeMetric.Usage.Cpu().MilliValue()
				node.Memory = toMegaBytes(nodeMetric.Usage.Memory().Value())
			}
		}
	}

	return nodeResources, nil
}

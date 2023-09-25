//nolint:gochecknoglobals
package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/raffepaffe/kload/internal/cmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Pod struct {
	Namespace  string
	Name       string
	Label      string
	Containers []*Container
}

type Container struct {
	Name        string
	CPU         int64
	CPULimit    int64
	Memory      int64
	MemoryLimit int64
}

var k8sSetup *client

func CreatePods(command *cmd.Command) ([]*Pod, error) {
	if k8sSetup == nil {
		s, err := setup(command)
		k8sSetup = s
		if err != nil {
			return nil, fmt.Errorf("setup error: %w", err)
		}
	}

	pods, err := createPods(k8sSetup, command)
	if err != nil {
		return nil, fmt.Errorf("createPods error: %w", err)
	}

	return pods, nil
}

func createPods(client *client, command *cmd.Command) ([]*Pod, error) {
	pods, err := createPodResources(client.ClientSet, command)
	if err != nil {
		return nil, fmt.Errorf("createPodResources error: %w", err)
	}

	mc, err := metrics.NewForConfig(client.Config)
	if err != nil {
		return nil, fmt.Errorf("NewForConfig error: %w", err)
	}

	pods, err = createPodMetrics(mc, command.Namespace, pods)
	if err != nil {
		return nil, fmt.Errorf("createPodMetrics error: %w", err)
	}

	return pods, nil
}

func createPodResources(clientset *kubernetes.Clientset, command *cmd.Command) ([]*Pod, error) {
	result := make([]*Pod, 0)
	pods, err := clientset.CoreV1().Pods(command.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return result, fmt.Errorf("pods error: %w", err)
	}

	for _, pod := range pods.Items {
		if keepPod(pod.Name, command.Attributes, command.ExcludePattern) {
			p := &Pod{
				Namespace: pod.Namespace,
				Name:      pod.Name,
				Label:     pod.Labels["app"],
			}
			for _, container := range pod.Spec.Containers {
				c := &Container{
					Name:        container.Name,
					CPULimit:    container.Resources.Limits.Cpu().MilliValue(),
					MemoryLimit: toMegaBytes(container.Resources.Limits.Memory().Value()),
				}
				p.Containers = append(p.Containers, c)
			}
			result = append(result, p)
		}
	}

	return result, nil
}

func createPodMetrics(mc *metrics.Clientset, namespace string, podResources []*Pod) ([]*Pod, error) {
	podMetrics, err := mc.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("PodMetricses error: %w", err)
	}

	for _, pod := range podResources {
		for _, podMetric := range podMetrics.Items {
			if pod.Name == podMetric.Name {
				for _, container := range pod.Containers {
					for _, containerMetric := range podMetric.Containers {
						if container.Name == containerMetric.Name {
							container.CPU = containerMetric.Usage.Cpu().MilliValue()
							container.Memory = toMegaBytes(containerMetric.Usage.Memory().Value())
						}
					}
				}
			}
		}
	}

	return podResources, nil
}

func keepPod(podName string, podNames []string, excludePattern string) bool {
	if len(podNames) == 0 {
		return true
	}
	for _, name := range podNames {
		if strings.HasPrefix(podName, name) {
			if utf8.RuneCountInString(excludePattern) == 0 || !matchPattern(excludePattern, podName) {
				return true
			}
		}
	}

	return false
}

func matchPattern(pattern, s string) bool {
	match, _ := regexp.MatchString(pattern, s)

	return match
}

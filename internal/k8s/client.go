package k8s

import (
	"fmt"

	"github.com/raffepaffe/kload/internal/cmd"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type client struct {
	Config    *rest.Config
	ClientSet *kubernetes.Clientset
}

// setup creates a kubernetes client.
func setup(command *cmd.Command) (*client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", command.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("build config error: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("new config error: %w", err)
	}

	return &client{
		Config:    config,
		ClientSet: clientset,
	}, nil
}

// toMegaBytes converts bytes to MB
// also, I don't care about the risk of overflow.
func toMegaBytes(n int64) int64 {
	f := float64(n) / (1 << 20)
	i := int64(f)

	return i
}

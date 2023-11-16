package cmd

import (
	"flag"
	"fmt"
	"path/filepath"
	"unicode/utf8"

	"k8s.io/client-go/util/homedir"
)

type Action int

const (
	Node Action = iota
	Pod
)

const defaultColumns = 3

type Command struct {
	Kubeconfig              string
	Action                  Action
	Namespace               string
	Attributes              []string
	MaxColumns              int
	ExcludePodPattern       string
	ExcludeContainerPattern string
}

// Parse creates a Command based in flags on the command line.
func Parse() (*Command, error) {
	nodeAction := flag.Bool("node", false, "show nodes")
	podAction := flag.Bool("pod", false, "show pods")
	namespace := flag.String("ns", "", "namespace to use with the -pod flag")
	maxColumns := flag.Int("columns", defaultColumns, "number of columns on one row")
	excludePod := flag.String("vp", "", "exclude pod name when used with the -pod flag")
	excludeContainer := flag.String("vc", "", "exclude container name when used with the -pod flag")

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	if !*nodeAction && !*podAction {
		flag.Usage()

		return nil, fmt.Errorf("expected '-node' or '-pod' flag")
	}

	if *podAction && utf8.RuneCountInString(*namespace) == 0 {
		flag.Usage()

		return nil, fmt.Errorf("expected '-ns' with a namespace name")
	}

	c := &Command{
		Kubeconfig:              *kubeconfig,
		Namespace:               *namespace,
		MaxColumns:              *maxColumns,
		ExcludePodPattern:       *excludePod,
		ExcludeContainerPattern: *excludeContainer,
	}

	if *nodeAction {
		c.Action = Node
	}

	if *podAction {
		c.Action = Pod
	}

	c.Attributes = flag.Args()

	return c, nil
}

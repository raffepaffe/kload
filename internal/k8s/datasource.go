package k8s

import (
	"fmt"

	"github.com/raffepaffe/kload/internal/cmd"
)

type DataSource struct {
	command  *cmd.Command
	elements []*Element
}

type Element struct {
	Name        string
	CPU         float64
	CPULimit    float64
	Memory      float64
	MemoryLimit float64
}

const aHundred = 100

func (e *Element) CPUPercent() float64 {
	return (e.CPU / e.CPULimit) * aHundred
}

func (e *Element) MemoryPercent() float64 {
	return (e.Memory / e.MemoryLimit) * aHundred
}

func (d *DataSource) Fetch() ([]*Element, error) {
	elements := make([]*Element, 0)
	if cmd.Node == d.command.Action {
		n, err := createNodes(d.command)
		if err != nil {
			return elements, fmt.Errorf("createNodesForClient error: %w", err)
		}
		d.elements = dataSourceFromNodes(n)
	}
	if cmd.Pod == d.command.Action {
		p, err := CreatePods(d.command)
		if err != nil {
			return elements, fmt.Errorf("createpods error: %w", err)
		}
		d.elements = datasourceFromPod(p)
	}

	return d.elements, nil
}

func (d *DataSource) MaxColumns() int {
	return d.command.MaxColumns
}

func NewDataSource(c *cmd.Command) (*DataSource, error) {
	ds := &DataSource{command: c}

	return ds, nil
}

func dataSourceFromNodes(nodes []*Node) []*Element {
	elements := make([]*Element, 0)
	for _, node := range nodes {
		e := &Element{
			Name:        node.Name,
			CPU:         float64(node.CPU),
			CPULimit:    float64(node.CPULimit),
			Memory:      float64(node.Memory),
			MemoryLimit: float64(node.MemoryLimit),
		}
		elements = append(elements, e)
	}

	return elements
}

func datasourceFromPod(pods []*Pod) []*Element {
	elements := make([]*Element, 0)
	for _, pod := range pods {
		for _, container := range pod.Containers {
			e := &Element{
				Name:        pod.Name + "/" + container.Name,
				CPU:         float64(container.CPU),
				CPULimit:    float64(container.CPULimit),
				Memory:      float64(container.Memory),
				MemoryLimit: float64(container.MemoryLimit),
			}
			elements = append(elements, e)
		}
	}

	return elements
}

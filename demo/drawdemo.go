package main

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/raffepaffe/kload/internal/k8s"
	"github.com/raffepaffe/kload/internal/ui/terminal"
)

type datasource struct {
	elements []*k8s.Element
}

func (d *datasource) Fetch() ([]*k8s.Element, error) {
	for _, e := range d.elements {
		value := randValue()
		e.CPU = e.CPU + float64(value)
		value = randValue()
		e.Memory = e.Memory + float64(value)
	}

	return d.elements, nil
}

func randValue() int32 {
	value := rand.Int31n(6)
	if value%2 == 0 {
		value = -value * 2
	}

	return value
}

func (d *datasource) MaxColumns() int {
	return 2
}

func newDatasource() *datasource {
	cpuLimit := float64(100)
	memLimit := float64(300)

	elements := make([]*k8s.Element, 0)

	for i := 0; i < 2; i++ {
		e := &k8s.Element{
			Name:        "http-server" + "/" + "agdf-1df" + strconv.Itoa(i),
			CPU:         rand.Float64() * 100,
			CPULimit:    cpuLimit,
			Memory:      rand.Float64() * 200,
			MemoryLimit: memLimit,
		}

		elements = append(elements, e)
	}

	ds := &datasource{elements: elements}

	return ds
}

func main() {
	fmt.Println("kload demo")
	ds := newDatasource()
	err := terminal.Draw(ds)
	if err != nil {
		fmt.Println("error is %w", err)
	}
}

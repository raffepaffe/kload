package main

import (
	"fmt"

	"github.com/raffepaffe/kload/internal/cmd"
	"github.com/raffepaffe/kload/internal/k8s"
	"github.com/raffepaffe/kload/internal/ui/terminal"
)

func main() {
	c, err := cmd.Parse()
	if err != nil {
		fmt.Println(err.Error())

		return
	}

	datasource, err := k8s.NewDataSource(c)
	if err != nil {
		fmt.Println(err.Error())

		return
	}

	err = terminal.Draw(datasource)
	if err != nil {
		fmt.Println(err.Error())
	}
}

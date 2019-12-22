package main

import (
	"os"

	"kubesphere.io/kubesphere/cmd/controller-manager/app"
)

func main() {
	command := app.NewControllerManagerCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

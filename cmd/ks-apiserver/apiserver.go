package main

import (
	"fmt"
	"kubesphere.io/kubesphere/cmd/ks-apiserver/app"
	"os"

	// Install apis
	_ "kubesphere.io/kubesphere/pkg/apis/metrics/install"
	_ "kubesphere.io/kubesphere/pkg/apis/operations/install"
	_ "kubesphere.io/kubesphere/pkg/apis/resources/install"
)

func main() {

	cmd := app.NewAPIServerCommand()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

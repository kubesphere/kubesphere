package controllers

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"kubesphere.io/kubesphere/pkg/resources"
)

func init() {
	flag.Set("logtostderr", "true")
}

func TestController(t *testing.T) {
	stopChan := make(chan struct{})
	resources.Sync(stopChan)
	Run(stopChan)

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	close(stopChan)
}

package app

import (
	"net/http"
	"testing"

	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/controllers"
	"kubesphere.io/kubesphere/pkg/resources"
)

func TestRun(t *testing.T) {
	stopChan := make(chan struct{})
	resources.Sync(stopChan)
	controllers.Run(stopChan)

	glog.Infoln("server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		glog.Infoln("exit with err:", err)
	}
	close(stopChan)

	glog.Infoln("server shutting down")
}

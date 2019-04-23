package sonarqube

import (
	"flag"
	"github.com/golang/glog"
	"github.com/kubesphere/sonargo/sonar"
	"strings"
	"sync"
)

var (
	sonarAddress string
	sonarToken   string
	sonarOnce    sync.Once
	sonarClient  *sonargo.Client
)

func init() {
	flag.StringVar(&sonarAddress, "sonar-address", "", "sonar server host")
	flag.StringVar(&sonarToken, "sonar-token", "", "sonar token")
}

func Client() *sonargo.Client {

	sonarOnce.Do(func() {
		if sonarAddress == "" {
			sonarClient = nil
			glog.Info("skip sonar init")
			return
		}
		if !strings.HasSuffix(sonarAddress, "/") {
			sonarAddress += "/"
		}
		client, err := sonargo.NewClientWithToken(sonarAddress+"api/", sonarToken)
		if err != nil {
			glog.Error("failed to connect to sonar")
			return
		}
		_, _, err = client.Projects.Search(nil)
		if err != nil {
			glog.Errorf("failed to search sonar projects [%+v]", err)
			return
		}
		glog.Info("init sonar client success")
		sonarClient = client
	})

	return sonarClient
}

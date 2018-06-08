/*
Copyright 2018 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cronjobs

import (
	"encoding/json"
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	"time"
)

var etcdClient *client.EtcdClient

var stopChan = make(chan struct{})

type dataType interface {
	namespace() string
}

type Worker interface {
	workOnce()
	chanRes() chan dataType
	chanStop() chan struct{}
}

func registerWorker(workers map[string]Worker, name string) {

	glog.Infof("Register cronjob: %s", name)
	k8sClient := client.NewK8sClient()
	switch name {
	case constants.WorkloadStatusKey:
		worker := workloadWorker{k8sClient: k8sClient, stopChan: stopChan, resChan: make(chan dataType, 10)}
		workers[constants.WorkloadStatusKey] = &worker
	case constants.QuotaKey:
		worker := resourceQuotaWorker{k8sClient: k8sClient, stopChan: stopChan, resChan: make(chan dataType, 10)}
		workers[constants.QuotaKey] = &worker
	}

}

func run(worker Worker) {

	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
			close(worker.chanRes())
		}
	}()

	for {
		select {
		case <-worker.chanStop():
			return
		default:
			break
		}

		worker.workOnce()
		time.Sleep(time.Duration(constants.UpdateCircle) * time.Second)

	}
}

func startWorks(workers map[string]Worker) {
	for wokername, woker := range workers {
		glog.Infof("cronjob %s start to work", wokername)
		go run(woker)
	}

}

func receiveResourceStatus(workers map[string]Worker) {
	defer func() {
		close(stopChan)
	}()

	for {
		for name, worker := range workers {
			select {
			case res, ok := <-worker.chanRes():
				if !ok {
					glog.Errorf("cronjob:%s have stopped", name)
					registerWorker(workers, name)
					run(workers[name])
				} else {
					value, err := json.Marshal(res)
					if err != nil {
						glog.Error(err)
						continue
					}
					key := constants.Root + "/" + name + "/" + res.namespace()
					err = etcdClient.Put(key, string(value))
					if err != nil {
						glog.Error(err)
					}
				}
			default:
				continue
			}
		}
	}
}

func Run() {
	glog.Info("Begin to run cronjob")
	var err error
	etcdClient, err = client.NewEtcdClient()
	if err != nil {
		glog.Error(err)
	}
	defer etcdClient.Close()
	workers := make(map[string]Worker)
	workerList := []string{constants.QuotaKey, constants.WorkloadStatusKey}
	for _, name := range workerList {
		registerWorker(workers, name)
	}
	startWorks(workers)
	receiveResourceStatus(workers)
}

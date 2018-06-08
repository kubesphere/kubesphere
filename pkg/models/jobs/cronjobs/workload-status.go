package cronjobs

import (
	"encoding/json"
	"time"

	"github.com/golang/glog"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
)

var workLoadList = []string{"deployments", "daemonsets", "statefulsets"}

type workLoadStatus struct {
	NameSpace       string
	Data            map[string]int
	UpdateTimeStamp int64
}

func (ws workLoadStatus) namespace() string {
	return ws.NameSpace
}

type workloadWorker struct {
	k8sClient *kubernetes.Clientset
	resChan   chan dataType
	stopChan  chan struct{}
}

func (ww *workloadWorker) GetNamespacesResourceStatus(namespace string) (map[string]int, error) {

	cli, err := client.NewEtcdClient()
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	defer cli.Close()

	res := make(map[string]int)

	for _, resourceName := range workLoadList {
		key := constants.Root + "/" + resourceName
		value, err := cli.Get(key)
		if err != nil {
			continue
		}

		resourceStatus := workload{ResourceList: make(workloadList)}

		err = json.Unmarshal(value, &resourceStatus)
		if err != nil {
			glog.Error(err)
			return nil, err
		}

		notReady := 0
		for _, v := range resourceStatus.ResourceList[namespace] {
			if !v.Ready {
				notReady++
			}
		}
		res[resourceName] = notReady
	}

	return res, nil
}

func (ww workloadWorker) workOnce() {
	namespaces, err := ww.k8sClient.CoreV1().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		glog.Error(err)
	}

	resourceStatus := make(map[string]int)
	for _, item := range namespaces.Items {
		namespace := item.Name
		namespacesResourceStatus, err := ww.GetNamespacesResourceStatus(namespace)
		if err != nil {
			glog.Error(err)
		}

		var ws = workLoadStatus{UpdateTimeStamp: time.Now().Unix(), Data: namespacesResourceStatus, NameSpace: namespace}
		ww.resChan <- ws

		for k, v := range namespacesResourceStatus {
			resourceStatus[k] = v + resourceStatus[k]
		}

	}

	var ws = workLoadStatus{UpdateTimeStamp: time.Now().Unix(), Data: resourceStatus, NameSpace: "\"\""}
	ww.resChan <- ws
}

func (ww workloadWorker) chanRes() chan dataType {
	return ww.resChan
}

func (ww workloadWorker) chanStop() chan struct{} {
	return ww.stopChan
}

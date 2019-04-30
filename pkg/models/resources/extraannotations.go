/*

 Copyright 2019 The KubeSphere Authors.

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
package resources

import (
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/informers"
)

type extraAnnotationInjector struct {
}

func (i extraAnnotationInjector) addExtraAnnotations(item interface{}) interface{} {

	switch item.(type) {
	case *v1.PersistentVolumeClaim:
		return i.injectPersistentVolumeClaim(item.(*v1.PersistentVolumeClaim))
	}

	return item
}

func (i extraAnnotationInjector) injectPersistentVolumeClaim(item *v1.PersistentVolumeClaim) *v1.PersistentVolumeClaim {
	podLister := informers.SharedInformerFactory().Core().V1().Pods().Lister()
	pods, err := podLister.Pods(item.Namespace).List(labels.Everything())
	if err != nil {
		glog.Errorf("inject annotation failed %+v", err)
		return item
	}

	item = item.DeepCopy()

	if item.Annotations == nil {
		item.Annotations = make(map[string]string, 0)
	}

	if isPvcInUse(pods, item.Name) {
		item.Annotations["kubesphere.io/in-use"] = "true"
	} else {
		item.Annotations["kubesphere.io/in-use"] = "false"
	}

	return item
}

func isPvcInUse(pods []*v1.Pod, pvcName string) bool {
	for _, pod := range pods {
		volumes := pod.Spec.Volumes
		for _, volume := range volumes {
			pvc := volume.PersistentVolumeClaim
			if pvc != nil && pvc.ClaimName == pvcName {
				return true
			}
		}
	}
	return false
}

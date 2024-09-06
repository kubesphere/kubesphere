/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package controller

import (
	"context"
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"kubesphere.io/kubesphere/pkg/controller/options"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
)

const (
	SyncFailed            = "SyncFailed"
	Synced                = "Synced"
	MessageResourceSynced = "Synced successfully"
)

type Controller interface {
	Name() string
	SetupWithManager(mgr *Manager) error
}

type Hideable interface {
	Hidden() bool
}

type ClusterSelector interface {
	Enabled(clusterRole string) bool
}

type Manager struct {
	options.Options
	manager.Manager
	K8sClient k8s.Client
	IsControllerEnabled
	ClusterClient clusterclient.Interface
	K8sVersion    *semver.Version
}

type IsControllerEnabled func(controllerName string) bool

func (mgr *Manager) Run(ctx context.Context, registry Registry) error {
	for name, ctr := range registry {
		if mgr.IsControllerEnabled(name) {
			if clusterSelector, ok := ctr.(ClusterSelector); ok &&
				!clusterSelector.Enabled(mgr.MultiClusterOptions.ClusterRole) {
				klog.Infof("%s controller is enabled but is not going to run due to its dependent component being disabled.", name)
				continue
			}
			if err := ctr.SetupWithManager(mgr); err != nil {
				return fmt.Errorf("unable to setup %s controller: %v", name, err)
			} else {
				klog.Infof("%s controller is enabled and added successfully.", name)
			}
		} else {
			klog.Infof("%s controller is disabled by controller selectors.", name)
		}
	}
	klog.V(0).Info("Starting the controllers.")
	if err := mgr.Manager.Start(ctx); err != nil {
		return fmt.Errorf("unable to start the controller manager: %v", err)
	}
	return nil
}

func Register(controller Controller) error {
	if _, exist := Controllers[controller.Name()]; exist {
		return fmt.Errorf("controller %s already exists", controller.Name())
	}
	Controllers[controller.Name()] = controller
	return nil
}

var Controllers = make(Registry)

type Registry map[string]Controller

func (r Registry) Keys() []string {
	var keys []string
	for k, v := range r {
		if hidden, ok := v.(Hideable); ok && hidden.Hidden() {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

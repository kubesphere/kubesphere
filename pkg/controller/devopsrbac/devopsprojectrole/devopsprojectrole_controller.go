package devopsprojectrole

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/metrics"
	"kubesphere.io/kubesphere/pkg/controller/devopsrbac"
	devopsmodel "kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"time"

	devopsv1alpha1 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	devopsclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	devopsinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/devops/v1alpha1"
	devopslisters "kubesphere.io/kubesphere/pkg/client/listers/devops/v1alpha1"
)

type DevopsProjectRoleController struct {
	client       clientset.Interface
	devopsClient devopsclient.Interface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	devopsProjectRoleListers devopslisters.DevOpsProjectRoleLister
	devopsProjectRoleSynced  cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration
}

func NewController(devopsclientset devopsclient.Interface,
	client clientset.Interface,
	devopsProjectRoleInformer devopsinformers.DevOpsProjectRoleInformer) *DevopsProjectRoleController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "devopsprojectrole-controller"})

	if client != nil && client.CoreV1().RESTClient().GetRateLimiter() != nil {
		metrics.RegisterMetricAndTrackRateLimiterUsage("devopsprojectrole-controller", client.CoreV1().RESTClient().GetRateLimiter())
	}

	v := &DevopsProjectRoleController{
		client:                   client,
		devopsClient:             devopsclientset,
		workqueue:                workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "devopsprojectrole"),
		devopsProjectRoleListers: devopsProjectRoleInformer.Lister(),
		devopsProjectRoleSynced:  devopsProjectRoleInformer.Informer().HasSynced,
		workerLoopPeriod:         time.Second,
	}

	v.eventBroadcaster = broadcaster
	v.eventRecorder = recorder

	devopsProjectRoleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: v.enqueueDevOpsProjectRole,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*devopsv1alpha1.DevOpsProjectRole)
			new := newObj.(*devopsv1alpha1.DevOpsProjectRole)
			if old.ResourceVersion == new.ResourceVersion {
				return
			}
			v.enqueueDevOpsProjectRole(newObj)
		},
		DeleteFunc: v.enqueueDevOpsProjectRole,
	})
	return v
}

// enqueueDevOpsProjectRole takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work workqueue. This method should *not* be
// passed resources of any type other than enqueueDevOpsProjectRole.
func (c *DevopsProjectRoleController) enqueueDevOpsProjectRole(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *DevopsProjectRoleController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.workqueue.Forget(obj)
		klog.V(5).Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		klog.Error(err, "could not reconcile DevOpsProjectRole")
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *DevopsProjectRoleController) worker() {

	for c.processNextWorkItem() {
	}
}

func (c *DevopsProjectRoleController) Start(stopCh <-chan struct{}) error {
	return c.Run(1, stopCh)
}

func (c *DevopsProjectRoleController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("starting DevOpsProjectRole controller")
	defer klog.Info("shutting down DevOpsProjectRole controller")

	if !cache.WaitForCacheSync(stopCh, c.devopsProjectRoleSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, c.workerLoopPeriod, stopCh)
	}

	<-stopCh
	return nil
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *DevopsProjectRoleController) syncHandler(key string) error {
	devopsProjectRole, err := c.devopsProjectRoleListers.Get(key)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Info(fmt.Sprintf("devopsprojectrole '%s' in work queue no longer exists ", key))
			return nil
		}
		klog.Error(err, fmt.Sprintf("could not get devopsprojectrole %s ", key))
		return err
	}
	if devopsProjectRole.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sliceutil.HasString(devopsProjectRole.ObjectMeta.Finalizers, devopsv1alpha1.DevOpsProjectRoleJenkinsFinalizerName) {
			devopsProjectRole.ObjectMeta.Finalizers = append(devopsProjectRole.ObjectMeta.Finalizers, devopsv1alpha1.DevOpsProjectRoleJenkinsFinalizerName)
			_, err := c.devopsClient.DevopsV1alpha1().DevOpsProjectRoles().Update(devopsProjectRole)
			if err != nil {
				klog.Error(err, fmt.Sprintf("failed to update devopsprojectrole %s ", key))
				return err
			}
		}

	} else {
		if sliceutil.HasString(devopsProjectRole.ObjectMeta.Finalizers, devopsv1alpha1.DevOpsProjectRoleJenkinsFinalizerName) {
			if err := c.deleteDevopsRoleInJenkins(devopsProjectRole); err != nil {
				klog.Error(err, fmt.Sprintf("failed to delete resource %s in jenkins", key))
				return err
			}
			devopsProjectRole.ObjectMeta.Finalizers = sliceutil.RemoveString(devopsProjectRole.ObjectMeta.Finalizers, func(item string) bool {
				return item == devopsv1alpha1.DevOpsProjectRoleJenkinsFinalizerName
			})
			_, err := c.devopsClient.DevopsV1alpha1().DevOpsProjectRoles().Update(devopsProjectRole)
			if err != nil {
				klog.Error(err, fmt.Sprintf("failed to update devopsprojectrole %s ", key))
				return err
			}
		}
	}

	exists, err := c.checkDevopsRoleExistsInJenkins(devopsProjectRole)
	if err != nil {
		klog.Error(err, fmt.Sprintf("failed to check resource %s exists in jenkins", key))
		return err
	}
	if exists {
		return nil
	}
	if err := c.addDevopsRoleInJenkins(devopsProjectRole); err != nil {
		klog.Error(err, fmt.Sprintf("failed to create resource %s in jenkins", key))
		return err
	}
	return nil
}

// checkDevopsRoleExistsInJenkins used to check if roles has been written to jenkins
func (c *DevopsProjectRoleController) checkDevopsRoleExistsInJenkins(role *devopsv1alpha1.DevOpsProjectRole) (bool, error) {
	devops, err := client.ClientSets().Devops()
	if err != nil {
		klog.Error(err)
		return false, err
	}

	roleType, err := devopsrbac.GetRoleTypeByRoleName(role.GetName())
	if err != nil {
		klog.Error(err)
		return false, err
	}
	prefix, err := devopsrbac.GetJenkinsRolePrefix(role)

	pipelineRoleName := devopsmodel.GetPipelineRoleName(prefix, roleType)
	projectRoleName := devopsmodel.GetProjectRoleName(prefix, roleType)

	projectRole, err := devops.Jenkins().GetProjectRole(projectRoleName)
	if err != nil {
		klog.Error(err)
		return false, err
	}
	pipelineRole, err := devops.Jenkins().GetProjectRole(pipelineRoleName)
	if err != nil {
		klog.Error(err)
		return false, err
	}
	if projectRole != nil && pipelineRole != nil {
		return true, nil
	} else {
		return false, nil
	}
}

// deleteDevopsRoleInJenkins used to delete roles in jenkins
func (c *DevopsProjectRoleController) deleteDevopsRoleInJenkins(role *devopsv1alpha1.DevOpsProjectRole) error {
	devops, err := client.ClientSets().Devops()
	if err != nil {
		klog.Error(err)
		return err
	}

	roleType, err := devopsrbac.GetRoleTypeByRoleName(role.GetName())
	if err != nil {
		klog.Error(err)
		return err
	}
	prefix, err := devopsrbac.GetJenkinsRolePrefix(role)
	pipelineRoleName := devopsmodel.GetPipelineRoleName(prefix, roleType)
	projectRoleName := devopsmodel.GetProjectRoleName(prefix, roleType)

	err = devops.Jenkins().DeleteProjectRoles(projectRoleName, pipelineRoleName)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

// addDevopsRoleInJenkins used to add roles to jenkins
func (c *DevopsProjectRoleController) addDevopsRoleInJenkins(role *devopsv1alpha1.DevOpsProjectRole) error {

	devops, err := client.ClientSets().Devops()
	if err != nil {
		klog.Error(err)
		return err
	}

	roleType, err := devopsrbac.GetRoleTypeByRoleName(role.GetName())
	if err != nil {
		klog.Error(err)
		return err
	}
	prefix, err := devopsrbac.GetJenkinsRolePrefix(role)
	pipelineRoleName := devopsmodel.GetPipelineRoleName(prefix, roleType)
	projectRoleName := devopsmodel.GetProjectRoleName(prefix, roleType)

	// Each devops project role corresponds to two roles in the jenkins role strategy.
	// A role is used to match "^workspace/devopsproject$" to indicate the availability of folders in jenkins
	// Another role is used to match "^workspace/devopsproject/.*" to illustrate the availability of the pipeline under the folder in jenkins
	_, err = devops.Jenkins().AddProjectRole(
		projectRoleName, devopsmodel.GetProjectRolePattern(prefix), devopsmodel.JenkinsProjectPermissionMap[roleType], true)
	if err != nil {
		klog.Error(err)
		return err
	}
	_, err = devops.Jenkins().AddProjectRole(
		pipelineRoleName, devopsmodel.GetPipelineRolePattern(prefix), devopsmodel.JenkinsPipelinePermissionMap[roleType], true)
	if err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

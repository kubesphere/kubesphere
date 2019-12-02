package devopsprojectrolebinding

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
	"kubesphere.io/kubesphere/pkg/gojenkins"
	devopsmodel "kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sync"
	"time"

	devopsv1alpha1 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha1"
	devopsclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	devopsinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/devops/v1alpha1"
	devopslisters "kubesphere.io/kubesphere/pkg/client/listers/devops/v1alpha1"
)

type DevopsProjectRoleBindingController struct {
	client       clientset.Interface
	devopsClient devopsclient.Interface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	devopsProjectRoleBindingListers devopslisters.DevOpsProjectRoleBindingLister
	devopsProjectRoleSynced         cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration
}

func NewController(devopsclientset devopsclient.Interface,
	client clientset.Interface,
	devopsProjectRoleBindingInformer devopsinformers.DevOpsProjectRoleBindingInformer) *DevopsProjectRoleBindingController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "devopsprojectrolebinding-controller"})

	if client != nil && client.CoreV1().RESTClient().GetRateLimiter() != nil {
		metrics.RegisterMetricAndTrackRateLimiterUsage("devopsprojectrolebinding-controller", client.CoreV1().RESTClient().GetRateLimiter())
	}

	v := &DevopsProjectRoleBindingController{
		client:                          client,
		devopsClient:                    devopsclientset,
		workqueue:                       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "devopsprojectrolebinding"),
		devopsProjectRoleBindingListers: devopsProjectRoleBindingInformer.Lister(),
		devopsProjectRoleSynced:         devopsProjectRoleBindingInformer.Informer().HasSynced,
		workerLoopPeriod:                time.Second,
	}

	v.eventBroadcaster = broadcaster
	v.eventRecorder = recorder

	devopsProjectRoleBindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: v.enqueueDevOpsProjectRoleBinding,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*devopsv1alpha1.DevOpsProjectRole)
			new := newObj.(*devopsv1alpha1.DevOpsProjectRole)
			if old.ResourceVersion == new.ResourceVersion {
				return
			}
			v.enqueueDevOpsProjectRoleBinding(newObj)
		},
		DeleteFunc: v.enqueueDevOpsProjectRoleBinding,
	})
	return v
}

// enqueueDevOpsProjectRoleBinding takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work workqueue. This method should *not* be
// passed resources of any type other than enqueueDevOpsProjectRoleBinding.
func (c *DevopsProjectRoleBindingController) enqueueDevOpsProjectRoleBinding(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *DevopsProjectRoleBindingController) processNextWorkItem() bool {
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
		klog.Error(err, "could not reconcile DevOpsProjectRoleBinding")
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *DevopsProjectRoleBindingController) worker() {

	for c.processNextWorkItem() {
	}
}

func (c *DevopsProjectRoleBindingController) Start(stopCh <-chan struct{}) error {
	return c.Run(1, stopCh)
}

func (c *DevopsProjectRoleBindingController) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("starting DevOpsProjectRoleBinding controller")
	defer klog.Info("shutting down DevOpsProjectRoleBinding controller")

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
func (c *DevopsProjectRoleBindingController) syncHandler(key string) error {
	devopsProjectRoleBinding, err := c.devopsProjectRoleBindingListers.Get(key)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Info(fmt.Sprintf("DevOpsProjectRoleBinding '%s' in work queue no longer exists ", key))
			return nil
		}
		klog.Error(err, fmt.Sprintf("could not get DevOpsProjectRoleBinding %s ", key))
		return err
	}

	exists, err := c.checkDevopsRoleExistsInJenkins(devopsProjectRoleBinding)
	if err != nil {
		klog.Error(err, fmt.Sprintf("failed to check resource %s exists in jenkins", key))
		return err
	}
	if !exists {
		c.workqueue.AddRateLimited(key)
		klog.Infof("role is has not been written into jenkins requeue %s", key)
		return nil
	}
	if err := c.reassignDevopsRoleInJenkins(devopsProjectRoleBinding); err != nil {
		klog.Error(err, fmt.Sprintf("failed to create resource %s in jenkins", key))
		return err
	}
	return nil
}

// checkDevopsRoleExistsInJenkins used to check if roles has been written to jenkins
func (c *DevopsProjectRoleBindingController) checkDevopsRoleExistsInJenkins(roleBinding *devopsv1alpha1.DevOpsProjectRoleBinding) (bool, error) {
	devops, err := client.ClientSets().Devops()
	if err != nil {
		klog.Error(err)
		return false, err
	}

	roleType, err := devopsrbac.GetRoleTypeByRoleName(roleBinding.Spec.RoleRef.Name)
	if err != nil {
		klog.Error(err)
		return false, err
	}
	prefix, err := devopsrbac.GetJenkinsRolePrefix(roleBinding)

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

// reassignDevopsRoleInJenkins used to detect differences between CRD spec and jenkins role strategy, and reallocate
func (c *DevopsProjectRoleBindingController) reassignDevopsRoleInJenkins(roleBinding *devopsv1alpha1.DevOpsProjectRoleBinding) error {

	devops, err := client.ClientSets().Devops()
	if err != nil {
		klog.Error(err)
		return err
	}

	roleType, err := devopsrbac.GetRoleTypeByRoleName(roleBinding.Spec.RoleRef.Name)
	if err != nil {
		klog.Error(err)
		return err
	}
	prefix, err := devopsrbac.GetJenkinsRolePrefix(roleBinding)
	pipelineRoleName := devopsmodel.GetPipelineRoleName(prefix, roleType)
	projectRoleName := devopsmodel.GetProjectRoleName(prefix, roleType)

	users := devopsrbac.RBACSubjectsToStringSlice(roleBinding.Spec.Subjects)

	projectRole, err := devops.Jenkins().GetProjectRole(projectRoleName)
	if err != nil {
		klog.Error(err)
		return err
	}
	pipelineRole, err := devops.Jenkins().GetProjectRole(pipelineRoleName)
	if err != nil {
		klog.Error(err)
		return err
	}
	shouldAssignProject, shouldUnassignProject := sliceutil.StringDiff(users, projectRole.Raw.Sids)
	shouldAssignPipeline, shouldUnassignPipeline := sliceutil.StringDiff(users, pipelineRole.Raw.Sids)
	var assignProjectRoleCh = make(chan error, len(shouldAssignProject))
	var assignProjectRoleWg sync.WaitGroup
	for _, username := range shouldAssignProject {
		assignProjectRoleWg.Add(1)
		go func(role *gojenkins.ProjectRole, sid string) {
			err := role.AssignRole(sid)
			assignProjectRoleCh <- err
			assignProjectRoleWg.Done()
		}(projectRole, username)
	}

	var unassignProjectRoleCh = make(chan error, len(shouldUnassignProject))
	var unassignProjectRoleWg sync.WaitGroup
	for _, username := range shouldUnassignProject {
		assignProjectRoleWg.Add(1)
		go func(role *gojenkins.ProjectRole, sid string) {
			err := role.UnAssignRole(sid)
			unassignProjectRoleCh <- err
			unassignProjectRoleWg.Done()
		}(projectRole, username)
	}

	var assignPipleineRoleCh = make(chan error, len(shouldAssignPipeline))
	var assignPipelineRoleWg sync.WaitGroup
	for _, username := range shouldAssignPipeline {
		assignProjectRoleWg.Add(1)
		go func(role *gojenkins.ProjectRole, sid string) {
			err := role.AssignRole(sid)
			assignPipleineRoleCh <- err
			assignPipelineRoleWg.Done()
		}(pipelineRole, username)
	}

	var unassignPipelineRoleCh = make(chan error, len(shouldUnassignPipeline))
	var unassignPipelineRoleWg sync.WaitGroup
	for _, username := range shouldUnassignPipeline {
		assignProjectRoleWg.Add(1)
		go func(role *gojenkins.ProjectRole, sid string) {
			err := role.UnAssignRole(sid)
			unassignPipelineRoleCh <- err
			unassignPipelineRoleWg.Done()
		}(pipelineRole, username)
	}

	assignProjectRoleWg.Wait()
	close(assignProjectRoleCh)

	unassignProjectRoleWg.Wait()
	close(unassignProjectRoleCh)

	assignPipelineRoleWg.Wait()
	close(assignPipleineRoleCh)

	unassignPipelineRoleWg.Wait()
	close(unassignPipelineRoleCh)
	for err := range assignProjectRoleCh {
		if err != nil {
			klog.Errorf("could not assign role %s", err)
			return err
		}
	}

	for err := range unassignProjectRoleCh {
		if err != nil {
			klog.Errorf("could not unassign role %s", err)
			return err
		}
	}

	for err := range assignPipleineRoleCh {
		if err != nil {
			klog.Errorf("could not assign role %s", err)
			return err
		}
	}

	for err := range unassignPipelineRoleCh {
		if err != nil {
			klog.Errorf("could not unassign role %s", err)
			return err
		}
	}

	return nil
}

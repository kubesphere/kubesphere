package devopsproject

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
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	devopsClient "kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"net/http"
	"time"

	kubesphereclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	devopsinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/devops/v1alpha3"
	devopslisters "kubesphere.io/kubesphere/pkg/client/listers/devops/v1alpha3"
)

type Controller struct {
	client           clientset.Interface
	kubesphereClient kubesphereclient.Interface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	devOpsProjectLister devopslisters.DevOpsProjectLister
	devOpsProjectSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration

	devopsClient devopsClient.Interface
}

func NewController(client clientset.Interface,
	kubesphereClient kubesphereclient.Interface,
	devopsClinet devopsClient.Interface,
	devopsInformer devopsinformers.DevOpsProjectInformer) *Controller {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: client.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "devopsproject-controller"})

	v := &Controller{
		client:              client,
		devopsClient:        devopsClinet,
		kubesphereClient:    kubesphereClient,
		workqueue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "devopsproject"),
		devOpsProjectLister: devopsInformer.Lister(),
		devOpsProjectSynced: devopsInformer.Informer().HasSynced,
		workerLoopPeriod:    time.Second,
	}

	v.eventBroadcaster = broadcaster
	v.eventRecorder = recorder

	devopsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: v.enqueueDevOpsProject,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*devopsv1alpha3.DevOpsProject)
			new := newObj.(*devopsv1alpha3.DevOpsProject)
			if old.ResourceVersion == new.ResourceVersion {
				return
			}
			v.enqueueDevOpsProject(newObj)
		},
		DeleteFunc: v.enqueueDevOpsProject,
	})
	return v
}

// enqueueDevOpsProject takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work workqueue. This method should *not* be
// passed resources of any type other than DevOpsProject.
func (c *Controller) enqueueDevOpsProject(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) processNextWorkItem() bool {
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
		klog.Error(err, "could not reconcile devopsProject")
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) worker() {

	for c.processNextWorkItem() {
	}
}

func (c *Controller) Start(stopCh <-chan struct{}) error {
	return c.Run(1, stopCh)
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("starting devops project controller")
	defer klog.Info("shutting down devops project controller")

	if !cache.WaitForCacheSync(stopCh, c.devOpsProjectSynced) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, c.workerLoopPeriod, stopCh)
	}

	<-stopCh
	return nil
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the devopsproject resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	project, err := c.devOpsProjectLister.Get(key)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.Info(fmt.Sprintf("devopsproject '%s' in work queue no longer exists ", key))
			return nil
		}
		klog.Error(err, fmt.Sprintf("could not get devopsproject %s ", key))
		return err
	}
	if project.ObjectMeta.DeletionTimestamp.IsZero() {
		if !sliceutil.HasString(project.ObjectMeta.Finalizers, devopsv1alpha3.DevOpsProjectFinalizerName) {
			project.ObjectMeta.Finalizers = append(project.ObjectMeta.Finalizers, devopsv1alpha3.DevOpsProjectFinalizerName)
			_, err := c.kubesphereClient.DevopsV1alpha3().DevOpsProjects().Update(project)
			if err != nil {
				klog.Error(err, fmt.Sprintf("failed to update project %s ", key))
				return err
			}
		}
		_, err := c.devopsClient.GetDevOpsProject(key)
		if err != nil && devopsClient.GetDevOpsStatusCode(err) != http.StatusNotFound {
			klog.Error(err, fmt.Sprintf("failed to get project %s ", key))
			return err
		} else {
			_, err := c.devopsClient.CreateDevOpsProject(key)
			if err != nil {
				klog.Error(err, fmt.Sprintf("failed to get project %s ", key))
				return err
			}
		}
	} else {
		if sliceutil.HasString(project.ObjectMeta.Finalizers, devopsv1alpha3.DevOpsProjectFinalizerName) {
			_, err := c.devopsClient.GetDevOpsProject(key)
			if err != nil && devopsClient.GetDevOpsStatusCode(err) != http.StatusNotFound {
				klog.Error(err, fmt.Sprintf("failed to get project %s ", key))
				return err
			} else if err != nil && devopsClient.GetDevOpsStatusCode(err) == http.StatusNotFound {
			} else {
				if err := c.deleteDevOpsProjectInDevOps(project); err != nil {
					klog.Error(err, fmt.Sprintf("failed to delete resource %s in devops", key))
					return err
				}
			}
			project.ObjectMeta.Finalizers = sliceutil.RemoveString(project.ObjectMeta.Finalizers, func(item string) bool {
				return item == devopsv1alpha3.DevOpsProjectFinalizerName
			})

			_, err = c.kubesphereClient.DevopsV1alpha3().DevOpsProjects().Update(project)
			if err != nil {
				klog.Error(err, fmt.Sprintf("failed to update project %s ", key))
				return err
			}
		}
	}

	return nil
}

func (c *Controller) deleteDevOpsProjectInDevOps(project *devopsv1alpha3.DevOpsProject) error {

	err := c.devopsClient.DeleteDevOpsProject(project.Name)
	if err != nil {
		klog.Errorf("error happened while deleting %s, %v", project.Name, err)
	}

	return nil
}

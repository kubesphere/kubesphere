package devopsproject

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1informer "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	devopsv1alpha3 "kubesphere.io/kubesphere/pkg/apis/devops/v1alpha3"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	"kubesphere.io/kubesphere/pkg/constants"
	devopsClient "kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/utils/k8sutil"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"net/http"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	kubesphereclient "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	devopsinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/devops/v1alpha3"
	devopslisters "kubesphere.io/kubesphere/pkg/client/listers/devops/v1alpha3"
)

/**
  DevOps project controller is used to maintain the state of the DevOps project.
*/

type Controller struct {
	client           clientset.Interface
	kubesphereClient kubesphereclient.Interface

	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	devOpsProjectLister devopslisters.DevOpsProjectLister
	devOpsProjectSynced cache.InformerSynced

	namespaceLister corev1lister.NamespaceLister
	namespaceSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	workerLoopPeriod time.Duration

	devopsClient devopsClient.Interface
}

func NewController(client clientset.Interface,
	kubesphereClient kubesphereclient.Interface,
	devopsClinet devopsClient.Interface,
	namespaceInformer corev1informer.NamespaceInformer,
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
		namespaceLister:     namespaceInformer.Lister(),
		namespaceSynced:     namespaceInformer.Informer().HasSynced,
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
			klog.Info(fmt.Sprintf("devops->devopsproject '%s' in work queue no longer exists ", key))
			return nil
		}
		klog.Errorf("devops->could not get devopsproject %s, %v", key, err)
		return err
	}
	copyProject := project.DeepCopy()
	// DeletionTimestamp.IsZero() means DevOps project has not been deleted.
	if project.ObjectMeta.DeletionTimestamp.IsZero() {
		// Use Finalizers to sync DevOps status when DevOps project was deleted
		// https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#finalizers
		if !sliceutil.HasString(project.ObjectMeta.Finalizers, devopsv1alpha3.DevOpsProjectFinalizerName) {
			copyProject.ObjectMeta.Finalizers = append(copyProject.ObjectMeta.Finalizers, devopsv1alpha3.DevOpsProjectFinalizerName)
		}

		if project.Status.AdminNamespace != "" {
			ns, err := c.namespaceLister.Get(project.Status.AdminNamespace)
			if err != nil && !errors.IsNotFound(err) {
				klog.Error(err, fmt.Sprintf("devops->faild to get namespace"))
				return err
			} else if errors.IsNotFound(err) {
				// if admin ns is not found, clean project status, rerun reconcile
				copyProject.Status.AdminNamespace = ""
				_, err := c.kubesphereClient.DevopsV1alpha3().DevOpsProjects().Update(copyProject)
				if err != nil {
					klog.Error(err, fmt.Sprintf("devops->failed to update project %s ", key))
					return err
				}
				c.enqueueDevOpsProject(key)
				return nil
			}
			// If ns exists, but the associated attributes with the project are not set correctly,
			// then reset the associated attributes
			if k8sutil.IsControlledBy(ns.OwnerReferences,
				devopsv1alpha3.ResourceKindDevOpsProject, project.Name) &&
				ns.Labels[constants.DevOpsProjectLabelKey] == project.Name {
			} else {
				copyNs := ns.DeepCopy()
				err := controllerutil.SetControllerReference(copyProject, copyNs, scheme.Scheme)
				if err != nil {
					klog.Error(err, fmt.Sprintf("devops->failed to set ownerreference %s ", key))
					return err
				}
				copyNs.Labels[constants.DevOpsProjectLabelKey] = project.Name
				_, err = c.client.CoreV1().Namespaces().Update(copyNs)
				if err != nil {
					klog.Error(err, fmt.Sprintf("devops->failed to update ns %s ", key))
					return err
				}
			}

		} else {
			// list ns by devops project
			namespaces, err := c.namespaceLister.List(
				labels.SelectorFromSet(labels.Set{constants.DevOpsProjectLabelKey: project.Name}))
			if err != nil {
				klog.Errorf("devops->failed to list ns %s, %v", key, err)
				return err
			}
			// if there is no ns, generate new one
			if len(namespaces) == 0 {
				ns := c.generateNewNamespace(project)
				ns, err := c.client.CoreV1().Namespaces().Create(ns)
				if err != nil {
					klog.Errorf("devops->failed to create ns %s, %v", key, err)
					return err
				}
				copyProject.Status.AdminNamespace = ns.Name
			} else if len(namespaces) != 0 {
				ns := namespaces[0]
				// reset ownerReferences
				if !k8sutil.IsControlledBy(ns.OwnerReferences,
					devopsv1alpha3.ResourceKindDevOpsProject, project.Name) {
					copyNs := ns.DeepCopy()
					err := controllerutil.SetControllerReference(copyProject, copyNs, scheme.Scheme)
					if err != nil {
						klog.Errorf("devops->failed to set ownerreference %s, %v", key, err)
						return err
					}
					copyNs.Labels[constants.DevOpsProjectLabelKey] = project.Name
					_, err = c.client.CoreV1().Namespaces().Update(copyNs)
					if err != nil {
						klog.Errorf("devops->failed to update ns %s, %v ", key, err)
						return err
					}
				}
				copyProject.Status.AdminNamespace = ns.Name
			}
		}

		if !reflect.DeepEqual(copyProject, project) {
			_, err := c.kubesphereClient.DevopsV1alpha3().DevOpsProjects().Update(copyProject)
			if err != nil {
				klog.Errorf("devops->failed to update ns %s, %v", key, err)
				return err
			}
		}
		// Check project exists, otherwise we will create it.
		_, err := c.devopsClient.GetDevOpsProject(copyProject.Status.AdminNamespace)
		if err != nil && devopsClient.GetDevOpsStatusCode(err) != http.StatusNotFound {
			klog.Errorf("devops->failed to get project %s, %v ", copyProject.Status.AdminNamespace, err)
			return err
		} else {
			_, err := c.devopsClient.CreateDevOpsProject(copyProject.Status.AdminNamespace)
			if err != nil {
				klog.Errorf("devops->failed to get project %s, %v ", copyProject.Status.AdminNamespace, err)
				return err
			}
		}

	} else {
		// Finalizers processing logic
		if sliceutil.HasString(project.ObjectMeta.Finalizers, devopsv1alpha3.DevOpsProjectFinalizerName) {
			_, err := c.devopsClient.GetDevOpsProject(key)
			if err != nil && devopsClient.GetDevOpsStatusCode(err) != http.StatusNotFound {
				klog.Errorf("devops->failed to get project %s, %v ", copyProject.Status.AdminNamespace, err)
				return err
			} else if err != nil && devopsClient.GetDevOpsStatusCode(err) == http.StatusNotFound {
			} else {
				if err := c.deleteDevOpsProjectInDevOps(project); err != nil {
					klog.Errorf("devops->failed to delete resource %s in devops, %v", key, err)
					return err
				}
			}
			project.ObjectMeta.Finalizers = sliceutil.RemoveString(project.ObjectMeta.Finalizers, func(item string) bool {
				return item == devopsv1alpha3.DevOpsProjectFinalizerName
			})

			_, err = c.kubesphereClient.DevopsV1alpha3().DevOpsProjects().Update(project)
			if err != nil {
				klog.Errorf("devops->failed to update project %s , %v", key, err)
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

func (c *Controller) generateNewNamespace(project *devopsv1alpha3.DevOpsProject) *v1.Namespace {
	ns := &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: project.Name,
			Labels:       map[string]string{constants.DevOpsProjectLabelKey: project.Name},
		},
	}
	controllerutil.SetControllerReference(project, ns, scheme.Scheme)
	return ns
}

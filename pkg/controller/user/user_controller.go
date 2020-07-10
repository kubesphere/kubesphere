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

package user

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	kubespherescheme "kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	iamv1alpha2informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/iam/v1alpha2"
	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	ldapclient "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"
	"time"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	successSynced = "Synced"
	// is synced successfully
	messageResourceSynced = "User synced successfully"
	controllerName        = "user-controller"
)

type Controller struct {
	k8sClient         kubernetes.Interface
	ksClient          kubesphere.Interface
	kubeconfig        kubeconfig.Interface
	userInformer      iamv1alpha2informers.UserInformer
	userLister        iamv1alpha2listers.UserLister
	userSynced        cache.InformerSynced
	cmSynced          cache.InformerSynced
	fedUserCache      cache.Store
	fedUserController cache.Controller
	ldapClient        ldapclient.Interface
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder            record.EventRecorder
	multiClusterEnabled bool
}

func NewController(k8sClient kubernetes.Interface, ksClient kubesphere.Interface,
	config *rest.Config, userInformer iamv1alpha2informers.UserInformer,
	fedUserCache cache.Store, fedUserController cache.Controller,
	configMapInformer corev1informers.ConfigMapInformer,
	ldapClient ldapclient.Interface, multiClusterEnabled bool) *Controller {
	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.

	utilruntime.Must(kubespherescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	var kubeconfigOperator kubeconfig.Interface
	if config != nil {
		kubeconfigOperator = kubeconfig.NewOperator(k8sClient, configMapInformer, config)
	}
	ctl := &Controller{
		k8sClient:           k8sClient,
		ksClient:            ksClient,
		kubeconfig:          kubeconfigOperator,
		userInformer:        userInformer,
		userLister:          userInformer.Lister(),
		userSynced:          userInformer.Informer().HasSynced,
		cmSynced:            configMapInformer.Informer().HasSynced,
		fedUserCache:        fedUserCache,
		fedUserController:   fedUserController,
		ldapClient:          ldapClient,
		workqueue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Users"),
		recorder:            recorder,
		multiClusterEnabled: multiClusterEnabled,
	}
	klog.Info("Setting up event handlers")
	userInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.enqueueUser,
		UpdateFunc: func(old, new interface{}) {
			ctl.enqueueUser(new)
		},
		DeleteFunc: ctl.enqueueUser,
	})
	return ctl
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting User controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")

	synced := make([]cache.InformerSynced, 0)
	synced = append(synced, c.userSynced, c.cmSynced)
	if c.multiClusterEnabled {
		synced = append(synced, c.fedUserController.HasSynced)
	}

	if ok := cache.WaitForCacheSync(stopCh, synced...); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")
	return nil
}

func (c *Controller) enqueueUser(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the reconcile, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.reconcile(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced %s:%s", "key", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) reconcile(key string) error {

	// Get the user with this name
	user, err := c.userLister.Get(key)
	if err != nil {
		// The user may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("user '%s' in work queue no longer exists", key))
			return nil
		}
		klog.Error(err)
		return err
	}

	// name of your custom finalizer
	finalizer := "finalizers.kubesphere.io/users"

	if user.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(user.Finalizers, finalizer) {
			user.ObjectMeta.Finalizers = append(user.ObjectMeta.Finalizers, finalizer)

			if user, err = c.ksClient.IamV1alpha2().Users().Update(user); err != nil {
				return err
			}
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(user.ObjectMeta.Finalizers, finalizer) {

			klog.V(4).Infof("delete user %s", key)
			if err = c.ldapClient.Delete(key); err != nil && err != ldapclient.ErrUserNotExists {
				klog.Error(err)
				return err
			}

			if err = c.deleteRoleBindings(user); err != nil {
				klog.Error(err)
				return err
			}

			// remove our finalizer from the list and update it.
			user.Finalizers = sliceutil.RemoveString(user.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})

			if _, err := c.ksClient.IamV1alpha2().Users().Update(user); err != nil {
				return err
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return nil
	}

	if err = c.ldapSync(user); err != nil {
		klog.Error(err)
		return err
	}

	if user, err = c.ensurePasswordIsEncrypted(user); err != nil {
		klog.Error(err)
		return err
	}

	if c.kubeconfig != nil {
		// ensure user kubeconfig configmap is created
		if err = c.kubeconfig.CreateKubeConfig(user); err != nil {
			klog.Error(err)
			return err
		}
	}

	// synchronization through kubefed-controller when multi cluster is enabled
	if c.multiClusterEnabled {
		if err = c.multiClusterSync(user); err != nil {
			klog.Error(err)
			return err
		}
	}

	c.recorder.Event(user, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

func (c *Controller) Start(stopCh <-chan struct{}) error {
	return c.Run(4, stopCh)
}

func (c *Controller) ensurePasswordIsEncrypted(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	encrypted, _ := strconv.ParseBool(user.Annotations[iamv1alpha2.PasswordEncryptedAnnotation])
	// password is not encrypted
	if !encrypted {
		password, err := encrypt(user.Spec.EncryptedPassword)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		user = user.DeepCopy()
		user.Spec.EncryptedPassword = password
		if user.Annotations == nil {
			user.Annotations = make(map[string]string, 0)
		}
		user.Annotations[iamv1alpha2.PasswordEncryptedAnnotation] = "true"
		user.Status.State = iamv1alpha2.UserActive
		return c.ksClient.IamV1alpha2().Users().Update(user)
	}

	return user, nil
}

func (c *Controller) ensureNotControlledByKubefed(user *iamv1alpha2.User) error {
	if user.Labels[constants.KubefedManagedLabel] != "false" {
		if user.Labels == nil {
			user.Labels = make(map[string]string, 0)
		}
		user = user.DeepCopy()
		user.Labels[constants.KubefedManagedLabel] = "false"
		_, err := c.ksClient.IamV1alpha2().Users().Update(user)
		if err != nil {
			klog.Error(err)
		}
	}
	return nil
}

func (c *Controller) multiClusterSync(user *iamv1alpha2.User) error {

	if err := c.ensureNotControlledByKubefed(user); err != nil {
		klog.Error(err)
		return err
	}

	obj, exist, err := c.fedUserCache.GetByKey(user.Name)
	if !exist {
		return c.createFederatedUser(user)
	}
	if err != nil {
		klog.Error(err)
		return err
	}

	var federatedUser iamv1alpha2.FederatedUser
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.(*unstructured.Unstructured).Object, &federatedUser); err != nil {
		klog.Error(err)
		return err
	}

	if !reflect.DeepEqual(federatedUser.Spec.Template.Spec, user.Spec) ||
		!reflect.DeepEqual(federatedUser.Spec.Template.Status, user.Status) ||
		!reflect.DeepEqual(federatedUser.Labels, user.Labels) ||
		!reflect.DeepEqual(federatedUser.Annotations, user.Annotations) {

		federatedUser.Labels = user.Labels
		federatedUser.Spec.Template.Spec = user.Spec
		federatedUser.Spec.Template.Status = user.Status
		federatedUser.Spec.Template.Labels = user.Labels
		federatedUser.Spec.Template.Annotations = user.Annotations
		return c.updateFederatedUser(&federatedUser)
	}

	return nil
}

func (c *Controller) createFederatedUser(user *iamv1alpha2.User) error {
	federatedUser := &iamv1alpha2.FederatedUser{
		TypeMeta: metav1.TypeMeta{
			Kind:       iamv1alpha2.FedUserKind,
			APIVersion: iamv1alpha2.FedUserResource.Group + "/" + iamv1alpha2.FedUserResource.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: user.Name,
		},
		Spec: iamv1alpha2.FederatedUserSpec{
			Template: iamv1alpha2.UserTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      user.Labels,
					Annotations: user.Annotations,
				},
				Spec:   user.Spec,
				Status: user.Status,
			},
			Placement: iamv1alpha2.Placement{
				ClusterSelector: iamv1alpha2.ClusterSelector{},
			},
		},
	}

	// must bind user lifecycle
	err := controllerutil.SetControllerReference(user, federatedUser, scheme.Scheme)
	if err != nil {
		return err
	}

	data, err := json.Marshal(federatedUser)
	if err != nil {
		return err
	}

	cli := c.k8sClient.(*kubernetes.Clientset)

	err = cli.RESTClient().Post().
		AbsPath(fmt.Sprintf("/apis/%s/%s/%s", iamv1alpha2.FedUserResource.Group,
			iamv1alpha2.FedUserResource.Version, iamv1alpha2.FedUserResource.Name)).
		Body(data).
		Do().Error()
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return nil
}

func (c *Controller) updateFederatedUser(fedUser *iamv1alpha2.FederatedUser) error {
	data, err := json.Marshal(fedUser)
	if err != nil {
		return err
	}

	cli := c.k8sClient.(*kubernetes.Clientset)

	err = cli.RESTClient().Put().
		AbsPath(fmt.Sprintf("/apis/%s/%s/%s/%s", iamv1alpha2.FedUserResource.Group,
			iamv1alpha2.FedUserResource.Version, iamv1alpha2.FedUserResource.Name, fedUser.Name)).
		Body(data).
		Do().Error()
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

func (c *Controller) ldapSync(user *iamv1alpha2.User) error {
	encrypted, _ := strconv.ParseBool(user.Annotations[iamv1alpha2.PasswordEncryptedAnnotation])
	if encrypted {
		return nil
	}
	_, err := c.ldapClient.Get(user.Name)
	if err != nil {
		if err == ldapclient.ErrUserNotExists {
			klog.V(4).Infof("create user %s", user.Name)
			return c.ldapClient.Create(user)
		}
		klog.Error(err)
		return err
	} else {
		klog.V(4).Infof("update user %s", user.Name)
		return c.ldapClient.Update(user)
	}
}

func (c *Controller) deleteRoleBindings(user *iamv1alpha2.User) error {
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{iamv1alpha2.UserReferenceLabel: user.Name}).String(),
	}
	deleteOptions := metav1.NewDeleteOptions(0)

	if err := c.ksClient.IamV1alpha2().GlobalRoleBindings().
		DeleteCollection(deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}

	if err := c.ksClient.IamV1alpha2().WorkspaceRoleBindings().
		DeleteCollection(deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}

	if err := c.k8sClient.RbacV1().ClusterRoleBindings().
		DeleteCollection(deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}

	if result, err := c.k8sClient.CoreV1().Namespaces().List(metav1.ListOptions{}); err != nil {
		klog.Error(err)
		return err
	} else {
		for _, namespace := range result.Items {
			if err := c.k8sClient.RbacV1().RoleBindings(namespace.Name).
				DeleteCollection(deleteOptions, listOptions); err != nil {
				klog.Error(err)
				return err
			}
		}
	}

	return nil
}

func encrypt(password string) (string, error) {
	// when user is already mapped to another identity, password is empty by default
	// unable to log in directly until password reset
	if password == "" {
		return "", nil
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

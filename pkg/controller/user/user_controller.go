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
	authoptions "kubesphere.io/kubesphere/pkg/apiserver/authentication/options"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	kubespherescheme "kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	iamv1alpha2informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions/iam/v1alpha2"
	iamv1alpha2listers "kubesphere.io/kubesphere/pkg/client/listers/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/constants"
	modelsdevops "kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
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

	// user finalizer
	finalizer = "finalizers.kubesphere.io/users"
)

type Controller struct {
	k8sClient         kubernetes.Interface
	ksClient          kubesphere.Interface
	kubeconfig        kubeconfig.Interface
	userLister        iamv1alpha2listers.UserLister
	userSynced        cache.InformerSynced
	loginRecordLister iamv1alpha2listers.LoginRecordLister
	loginRecordSynced cache.InformerSynced
	cmSynced          cache.InformerSynced
	fedUserCache      cache.Store
	fedUserController cache.Controller
	ldapClient        ldapclient.Interface
	devopsClient      devops.Interface
	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder              record.EventRecorder
	authenticationOptions *authoptions.AuthenticationOptions
	multiClusterEnabled   bool
}

func NewUserController(k8sClient kubernetes.Interface, ksClient kubesphere.Interface,
	config *rest.Config, userInformer iamv1alpha2informers.UserInformer,
	fedUserCache cache.Store, fedUserController cache.Controller,
	loginRecordInformer iamv1alpha2informers.LoginRecordInformer,
	configMapInformer corev1informers.ConfigMapInformer,
	ldapClient ldapclient.Interface,
	devopsClient devops.Interface,
	authenticationOptions *authoptions.AuthenticationOptions,
	multiClusterEnabled bool) *Controller {

	utilruntime.Must(kubespherescheme.AddToScheme(scheme.Scheme))

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	var kubeconfigOperator kubeconfig.Interface
	if config != nil {
		kubeconfigOperator = kubeconfig.NewOperator(k8sClient, configMapInformer, config)
	}
	ctl := &Controller{
		k8sClient:             k8sClient,
		ksClient:              ksClient,
		kubeconfig:            kubeconfigOperator,
		userLister:            userInformer.Lister(),
		userSynced:            userInformer.Informer().HasSynced,
		loginRecordLister:     loginRecordInformer.Lister(),
		loginRecordSynced:     loginRecordInformer.Informer().HasSynced,
		cmSynced:              configMapInformer.Informer().HasSynced,
		fedUserCache:          fedUserCache,
		fedUserController:     fedUserController,
		ldapClient:            ldapClient,
		devopsClient:          devopsClient,
		workqueue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Users"),
		recorder:              recorder,
		multiClusterEnabled:   multiClusterEnabled,
		authenticationOptions: authenticationOptions,
	}

	klog.Info("Setting up event handlers")
	userInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.enqueueUser,
		UpdateFunc: func(old, new interface{}) {
			ctl.enqueueUser(new)
		},
		DeleteFunc: ctl.enqueueUser,
	})

	loginRecordInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(new interface{}) {
			if err := ctl.enqueueLogin(new); err != nil {
				klog.Errorf("Failed to enqueue login object, error: %v", err)
			}
		},
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
	synced = append(synced, c.userSynced, c.loginRecordSynced, c.cmSynced)
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

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.reconcile(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
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

// enqueueLogin accepts a login object and set user lastLoginTime field
func (c *Controller) enqueueLogin(object interface{}) error {
	login := object.(*iamv1alpha2.LoginRecord)
	username, ok := login.Labels[iamv1alpha2.UserReferenceLabel]

	if !ok || len(username) == 0 {
		return fmt.Errorf("login doesn't belong to any user")
	}

	user, err := c.userLister.Get(username)
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("user %s doesn't exist any more, login record will be deleted later", username)
		}
		return err
	}

	if user.Status.LastLoginTime == nil || user.Status.LastLoginTime.Before(&login.CreationTimestamp) {
		user.Status.LastLoginTime = &login.CreationTimestamp
		user, err = c.ksClient.IamV1alpha2().Users().Update(user)
		return err
	}

	return nil
}

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

			if c.devopsClient != nil {
				// unassign jenkins role, unassign multiple times is allowed
				if err := c.unassignDevOpsAdminRole(user); err != nil {
					klog.Error(err)
					return err
				}
			}

			if err = c.deleteLoginRecords(user); err != nil {
				klog.Error(err)
				return err
			}

			// remove our finalizer from the list and update it.
			user.Finalizers = sliceutil.RemoveString(user.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})

			if user, err = c.ksClient.IamV1alpha2().Users().Update(user); err != nil {
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

	if user, err = c.syncUserStatus(user); err != nil {
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

	if c.devopsClient != nil {
		// assign jenkins role after user create, assign multiple times is allowed
		// used as logged-in users can do anything
		if err := c.assignDevOpsAdminRole(user); err != nil {
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
	return c.Run(5, stopCh)
}

func (c *Controller) ensurePasswordIsEncrypted(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	encrypted := user.Annotations[iamv1alpha2.PasswordEncryptedAnnotation] == "true"
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
		// ensure plain text password won't be kept anywhere
		delete(user.Annotations, corev1.LastAppliedConfigAnnotation)
		user.Annotations[iamv1alpha2.PasswordEncryptedAnnotation] = "true"
		user.Status = iamv1alpha2.UserStatus{
			State:              iamv1alpha2.UserActive,
			LastTransitionTime: &metav1.Time{Time: time.Now()},
		}
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
		!reflect.DeepEqual(federatedUser.Spec.Template.Status, user.Status) {

		federatedUser.Spec.Template.Spec = user.Spec
		federatedUser.Spec.Template.Status = user.Status
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

func (c *Controller) assignDevOpsAdminRole(user *iamv1alpha2.User) error {
	if err := c.devopsClient.AssignGlobalRole(modelsdevops.JenkinsAdminRoleName, user.Name); err != nil {
		klog.Errorf("%+v", err)
		return err
	}
	return nil
}

func (c *Controller) unassignDevOpsAdminRole(user *iamv1alpha2.User) error {
	if err := c.devopsClient.UnAssignGlobalRole(modelsdevops.JenkinsAdminRoleName, user.Name); err != nil {
		klog.Errorf("%+v", err)
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
			if err = c.k8sClient.RbacV1().RoleBindings(namespace.Name).DeleteCollection(deleteOptions, listOptions); err != nil {
				klog.Error(err)
				return err
			}
		}
	}

	return nil
}

func (c *Controller) deleteLoginRecords(user *iamv1alpha2.User) error {
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{iamv1alpha2.UserReferenceLabel: user.Name}).String(),
	}
	deleteOptions := metav1.NewDeleteOptions(0)

	if err := c.ksClient.IamV1alpha2().LoginRecords().
		DeleteCollection(deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

// syncUserStatus will reconcile user state based on user login records
func (c *Controller) syncUserStatus(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	// disabled user, nothing to do
	if user == nil || (user.Status.State == iamv1alpha2.UserDisabled) {
		return user, nil
	}

	// blocked user, check if need to unblock user
	if user.Status.State == iamv1alpha2.UserAuthLimitExceeded {
		if user.Status.LastTransitionTime != nil &&
			user.Status.LastTransitionTime.Add(c.authenticationOptions.AuthenticateRateLimiterDuration).Before(time.Now()) {
			expected := user.DeepCopy()
			// unblock user
			if user.Annotations[iamv1alpha2.PasswordEncryptedAnnotation] == "true" {
				expected.Status = iamv1alpha2.UserStatus{
					State:              iamv1alpha2.UserActive,
					LastTransitionTime: &metav1.Time{Time: time.Now()},
				}
			}

			if !reflect.DeepEqual(expected.Status, user.Status) {
				return c.ksClient.IamV1alpha2().Users().Update(expected)
			}
		}
	}

	// normal user, check user's login records see if we need to block
	records, err := c.loginRecordLister.List(labels.SelectorFromSet(labels.Set{iamv1alpha2.UserReferenceLabel: user.Name}))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	// count failed login attempts during last AuthenticateRateLimiterDuration
	now := time.Now()
	failedLoginAttempts := 0
	for _, loginRecord := range records {
		if !loginRecord.Spec.Success && loginRecord.CreationTimestamp.Add(c.authenticationOptions.AuthenticateRateLimiterDuration).After(now) {
			failedLoginAttempts++
		}
	}

	// block user if failed login attempts exceeds maximum tries setting
	if failedLoginAttempts >= c.authenticationOptions.AuthenticateRateLimiterMaxTries {
		expect := user.DeepCopy()
		expect.Status = iamv1alpha2.UserStatus{
			State:              iamv1alpha2.UserAuthLimitExceeded,
			Reason:             fmt.Sprintf("Failed login attempts exceed %d in last %s", failedLoginAttempts, c.authenticationOptions.AuthenticateRateLimiterDuration),
			LastTransitionTime: &metav1.Time{Time: time.Now()},
		}

		// block user for AuthenticateRateLimiterDuration duration, after that put it back to the queue to unblock
		c.workqueue.AddAfter(user.Name, c.authenticationOptions.AuthenticateRateLimiterDuration)

		return c.ksClient.IamV1alpha2().Users().Update(expect)
	}
	return user, nil
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

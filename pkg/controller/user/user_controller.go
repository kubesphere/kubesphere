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
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/util/validation"

	utilwait "k8s.io/apimachinery/pkg/util/wait"

	"kubesphere.io/kubesphere/pkg/controller/utils/controller"

	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	iamv1alpha2 "kubesphere.io/api/iam/v1alpha2"

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
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	successSynced = "Synced"
	// is synced successfully
	messageResourceSynced = "User synced successfully"
	controllerName        = "user-controller"
	// user finalizer
	finalizer       = "finalizers.kubesphere.io/users"
	interval        = time.Second
	timeout         = 15 * time.Second
	syncFailMessage = "Failed to sync: %s"
)

type userController struct {
	controller.BaseController
	k8sClient             kubernetes.Interface
	ksClient              kubesphere.Interface
	kubeconfig            kubeconfig.Interface
	userLister            iamv1alpha2listers.UserLister
	loginRecordLister     iamv1alpha2listers.LoginRecordLister
	fedUserCache          cache.Store
	ldapClient            ldapclient.Interface
	devopsClient          devops.Interface
	authenticationOptions *authoptions.AuthenticationOptions
	multiClusterEnabled   bool
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

func NewUserController(k8sClient kubernetes.Interface, ksClient kubesphere.Interface, config *rest.Config,
	userInformer iamv1alpha2informers.UserInformer,
	loginRecordInformer iamv1alpha2informers.LoginRecordInformer,
	fedUserCache cache.Store, fedUserController cache.Controller,
	configMapInformer corev1informers.ConfigMapInformer,
	ldapClient ldapclient.Interface,
	devopsClient devops.Interface,
	authenticationOptions *authoptions.AuthenticationOptions,
	multiClusterEnabled bool) *userController {

	utilruntime.Must(kubespherescheme.AddToScheme(scheme.Scheme))

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: k8sClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})
	var kubeconfigOperator kubeconfig.Interface
	if config != nil {
		kubeconfigOperator = kubeconfig.NewOperator(k8sClient, configMapInformer, config)
	}
	ctl := &userController{
		BaseController: controller.BaseController{
			Workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "User"),
			Synced: []cache.InformerSynced{
				userInformer.Informer().HasSynced,
				configMapInformer.Informer().HasSynced,
				loginRecordInformer.Informer().HasSynced,
			},
			Name: controllerName,
		},
		k8sClient:             k8sClient,
		ksClient:              ksClient,
		kubeconfig:            kubeconfigOperator,
		userLister:            userInformer.Lister(),
		loginRecordLister:     loginRecordInformer.Lister(),
		fedUserCache:          fedUserCache,
		ldapClient:            ldapClient,
		devopsClient:          devopsClient,
		recorder:              recorder,
		multiClusterEnabled:   multiClusterEnabled,
		authenticationOptions: authenticationOptions,
	}
	if multiClusterEnabled {
		ctl.Synced = append(ctl.Synced, fedUserController.HasSynced)
	}
	ctl.Handler = ctl.reconcile
	klog.Info("Setting up event handlers")
	userInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctl.Enqueue,
		UpdateFunc: func(old, new interface{}) {
			ctl.Enqueue(new)
		},
		DeleteFunc: ctl.Enqueue,
	})
	return ctl
}

func (c *userController) Start(stopCh <-chan struct{}) error {
	return c.Run(5, stopCh)
}

func (c *userController) reconcile(key string) error {
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
			if user, err = c.ksClient.IamV1alpha2().Users().Update(context.Background(), user, metav1.UpdateOptions{}); err != nil {
				klog.Error(err)
				return err
			}
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(user.ObjectMeta.Finalizers, finalizer) {
			// we do not need to delete the user from ldapServer when ldapClient is nil
			if c.ldapClient != nil {
				if err = c.waitForDeleteFromLDAP(key); err != nil {
					// ignore timeout error
					c.recorder.Event(user, corev1.EventTypeWarning, controller.FailedSynced, fmt.Sprintf(syncFailMessage, err))
				}
			}

			if err = c.deleteRoleBindings(user); err != nil {
				c.recorder.Event(user, corev1.EventTypeWarning, controller.FailedSynced, fmt.Sprintf(syncFailMessage, err))
				return err
			}

			if err = c.deleteGroupBindings(user); err != nil {
				c.recorder.Event(user, corev1.EventTypeWarning, controller.FailedSynced, fmt.Sprintf(syncFailMessage, err))
				return err
			}

			if c.devopsClient != nil {
				// unassign jenkins role, unassign multiple times is allowed
				if err = c.waitForUnassignDevOpsAdminRole(user); err != nil {
					// ignore timeout error
					c.recorder.Event(user, corev1.EventTypeWarning, controller.FailedSynced, fmt.Sprintf(syncFailMessage, err))
				}
			}

			if err = c.deleteLoginRecords(user); err != nil {
				c.recorder.Event(user, corev1.EventTypeWarning, controller.FailedSynced, fmt.Sprintf(syncFailMessage, err))
				return err
			}

			// remove our finalizer from the list and update it.
			user.Finalizers = sliceutil.RemoveString(user.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})

			if user, err = c.ksClient.IamV1alpha2().Users().Update(context.Background(), user, metav1.UpdateOptions{}); err != nil {
				klog.Error(err)
				return err
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return nil
	}

	// we do not need to sync ldap info when ldapClient is nil
	if c.ldapClient != nil {
		// ignore errors if timeout
		if err = c.waitForSyncToLDAP(user); err != nil {
			// ignore timeout error
			c.recorder.Event(user, corev1.EventTypeWarning, controller.FailedSynced, fmt.Sprintf(syncFailMessage, err))
		}
	}

	if user, err = c.encryptPassword(user); err != nil {
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
			c.recorder.Event(user, corev1.EventTypeWarning, controller.FailedSynced, fmt.Sprintf(syncFailMessage, err))
			return err
		}
	}

	if c.devopsClient != nil {
		// assign jenkins role after user create, assign multiple times is allowed
		// used as logged-in users can do anything
		if err = c.waitForAssignDevOpsAdminRole(user); err != nil {
			// ignore timeout error
			c.recorder.Event(user, corev1.EventTypeWarning, controller.FailedSynced, fmt.Sprintf(syncFailMessage, err))
		}
	}

	// synchronization through kubefed-controller when multi cluster is enabled
	if c.multiClusterEnabled {
		if err = c.multiClusterSync(user); err != nil {
			c.recorder.Event(user, corev1.EventTypeWarning, controller.FailedSynced, fmt.Sprintf(syncFailMessage, err))
			return err
		}
	}

	c.recorder.Event(user, corev1.EventTypeNormal, successSynced, messageResourceSynced)
	return nil
}

func (c *userController) encryptPassword(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {
	// password is not empty and not encrypted
	if user.Spec.EncryptedPassword != "" && !isEncrypted(user.Spec.EncryptedPassword) {
		password, err := encrypt(user.Spec.EncryptedPassword)
		if err != nil {
			klog.Error(err)
			return nil, err
		}
		user = user.DeepCopy()
		user.Spec.EncryptedPassword = password
		if user.Annotations == nil {
			user.Annotations = make(map[string]string)
		}
		user.Annotations[iamv1alpha2.LastPasswordChangeTimeAnnotation] = time.Now().UTC().Format(time.RFC3339)
		// ensure plain text password won't be kept anywhere
		delete(user.Annotations, corev1.LastAppliedConfigAnnotation)
		return c.ksClient.IamV1alpha2().Users().Update(context.Background(), user, metav1.UpdateOptions{})
	}
	return user, nil
}

func (c *userController) ensureNotControlledByKubefed(user *iamv1alpha2.User) error {
	if user.Labels[constants.KubefedManagedLabel] != "false" {
		if user.Labels == nil {
			user.Labels = make(map[string]string, 0)
		}
		user = user.DeepCopy()
		user.Labels[constants.KubefedManagedLabel] = "false"
		_, err := c.ksClient.IamV1alpha2().Users().Update(context.Background(), user, metav1.UpdateOptions{})
		if err != nil {
			klog.Error(err)
		}
	}
	return nil
}

func (c *userController) multiClusterSync(user *iamv1alpha2.User) error {
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

func (c *userController) createFederatedUser(user *iamv1alpha2.User) error {
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
		Do(context.Background()).Error()
	if err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}

	return nil
}

func (c *userController) updateFederatedUser(fedUser *iamv1alpha2.FederatedUser) error {
	data, err := json.Marshal(fedUser)
	if err != nil {
		return err
	}

	cli := c.k8sClient.(*kubernetes.Clientset)
	err = cli.RESTClient().Put().
		AbsPath(fmt.Sprintf("/apis/%s/%s/%s/%s", iamv1alpha2.FedUserResource.Group,
			iamv1alpha2.FedUserResource.Version, iamv1alpha2.FedUserResource.Name, fedUser.Name)).
		Body(data).
		Do(context.Background()).Error()
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
		return err
	}
	return nil
}

func (c *userController) waitForAssignDevOpsAdminRole(user *iamv1alpha2.User) error {
	err := utilwait.PollImmediate(interval, timeout, func() (done bool, err error) {
		if err := c.devopsClient.AssignGlobalRole(modelsdevops.JenkinsAdminRoleName, user.Name); err != nil {
			klog.Error(err)
			return false, err
		}
		return true, nil
	})
	return err
}

func (c *userController) waitForUnassignDevOpsAdminRole(user *iamv1alpha2.User) error {
	err := utilwait.PollImmediate(interval, timeout, func() (done bool, err error) {
		if err := c.devopsClient.UnAssignGlobalRole(modelsdevops.JenkinsAdminRoleName, user.Name); err != nil {
			return false, err
		}
		return true, nil
	})
	return err
}

func (c *userController) waitForSyncToLDAP(user *iamv1alpha2.User) error {
	if isEncrypted(user.Spec.EncryptedPassword) {
		return nil
	}
	err := utilwait.PollImmediate(interval, timeout, func() (done bool, err error) {
		_, err = c.ldapClient.Get(user.Name)
		if err != nil {
			if err == ldapclient.ErrUserNotExists {
				err = c.ldapClient.Create(user)
				if err != nil {
					klog.Error(err)
					return false, err
				}
				return true, nil
			}
			klog.Error(err)
			return false, err
		}
		err = c.ldapClient.Update(user)
		if err != nil {
			klog.Error(err)
			return false, err
		}
		return true, nil
	})
	return err
}

func (c *userController) waitForDeleteFromLDAP(username string) error {
	err := utilwait.PollImmediate(interval, timeout, func() (done bool, err error) {
		err = c.ldapClient.Delete(username)
		if err != nil && err != ldapclient.ErrUserNotExists {
			klog.Error(err)
			return false, err
		}
		return true, nil
	})
	return err
}

func (c *userController) deleteGroupBindings(user *iamv1alpha2.User) error {
	// Groupbindings that created by kubeshpere will be deleted directly.
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{iamv1alpha2.UserReferenceLabel: user.Name}).String(),
	}
	if err := c.ksClient.IamV1alpha2().GroupBindings().
		DeleteCollection(context.Background(), *metav1.NewDeleteOptions(0), listOptions); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (c *userController) deleteRoleBindings(user *iamv1alpha2.User) error {
	if len(user.Name) > validation.LabelValueMaxLength {
		// ignore invalid label value error
		return nil
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(labels.Set{iamv1alpha2.UserReferenceLabel: user.Name}).String(),
	}
	deleteOptions := *metav1.NewDeleteOptions(0)
	if err := c.ksClient.IamV1alpha2().GlobalRoleBindings().
		DeleteCollection(context.Background(), deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}

	if err := c.ksClient.IamV1alpha2().WorkspaceRoleBindings().
		DeleteCollection(context.Background(), deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}

	if err := c.k8sClient.RbacV1().ClusterRoleBindings().
		DeleteCollection(context.Background(), deleteOptions, listOptions); err != nil {
		klog.Error(err)
		return err
	}

	if result, err := c.k8sClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{}); err != nil {
		klog.Error(err)
		return err
	} else {
		for _, namespace := range result.Items {
			if err = c.k8sClient.RbacV1().RoleBindings(namespace.Name).DeleteCollection(context.Background(), deleteOptions, listOptions); err != nil {
				klog.Error(err)
				return err
			}
		}
	}
	return nil
}

func (c *userController) deleteLoginRecords(user *iamv1alpha2.User) error {
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{iamv1alpha2.UserReferenceLabel: user.Name}).String(),
	}
	if err := c.ksClient.IamV1alpha2().LoginRecords().
		DeleteCollection(context.Background(), *metav1.NewDeleteOptions(0), listOptions); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

// syncUserStatus will reconcile user state based on user login records
func (c *userController) syncUserStatus(user *iamv1alpha2.User) (*iamv1alpha2.User, error) {

	if user.Spec.EncryptedPassword == "" {
		if user.Labels[iamv1alpha2.IdentifyProviderLabel] != "" {
			// mapped user from other identity provider always active until disabled
			if user.Status.State == nil || *user.Status.State != iamv1alpha2.UserActive {
				expected := user.DeepCopy()
				active := iamv1alpha2.UserActive
				expected.Status = iamv1alpha2.UserStatus{
					State:              &active,
					LastTransitionTime: &metav1.Time{Time: time.Now()},
				}
				return c.ksClient.IamV1alpha2().Users().Update(context.Background(), expected, metav1.UpdateOptions{})
			}
		} else {
			// becomes disabled after setting a blank password
			if user.Status.State == nil || *user.Status.State != iamv1alpha2.UserDisabled {
				expected := user.DeepCopy()
				disabled := iamv1alpha2.UserDisabled
				expected.Status = iamv1alpha2.UserStatus{
					State:              &disabled,
					LastTransitionTime: &metav1.Time{Time: time.Now()},
				}
				return c.ksClient.IamV1alpha2().Users().Update(context.Background(), expected, metav1.UpdateOptions{})
			}
		}
		return user, nil
	}

	// becomes active after password encrypted
	if isEncrypted(user.Spec.EncryptedPassword) {
		if user.Status.State == nil || *user.Status.State == iamv1alpha2.UserDisabled {
			expected := user.DeepCopy()
			active := iamv1alpha2.UserActive
			expected.Status = iamv1alpha2.UserStatus{
				State:              &active,
				LastTransitionTime: &metav1.Time{Time: time.Now()},
			}
			return c.ksClient.IamV1alpha2().Users().Update(context.Background(), expected, metav1.UpdateOptions{})
		}
	}

	// blocked user, check if need to unblock user
	if user.Status.State != nil && *user.Status.State == iamv1alpha2.UserAuthLimitExceeded {
		if user.Status.LastTransitionTime != nil &&
			user.Status.LastTransitionTime.Add(c.authenticationOptions.AuthenticateRateLimiterDuration).Before(time.Now()) {
			expected := user.DeepCopy()
			// unblock user
			active := iamv1alpha2.UserActive
			expected.Status = iamv1alpha2.UserStatus{
				State:              &active,
				LastTransitionTime: &metav1.Time{Time: time.Now()},
			}
			return c.ksClient.IamV1alpha2().Users().Update(context.Background(), expected, metav1.UpdateOptions{})
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
		if !loginRecord.Spec.Success &&
			loginRecord.CreationTimestamp.Add(c.authenticationOptions.AuthenticateRateLimiterDuration).After(now) {
			failedLoginAttempts++
		}
	}

	// block user if failed login attempts exceeds maximum tries setting
	if failedLoginAttempts >= c.authenticationOptions.AuthenticateRateLimiterMaxTries {
		expect := user.DeepCopy()
		limitExceed := iamv1alpha2.UserAuthLimitExceeded
		expect.Status = iamv1alpha2.UserStatus{
			State:              &limitExceed,
			Reason:             fmt.Sprintf("Failed login attempts exceed %d in last %s", failedLoginAttempts, c.authenticationOptions.AuthenticateRateLimiterDuration),
			LastTransitionTime: &metav1.Time{Time: time.Now()},
		}
		// block user for AuthenticateRateLimiterDuration duration, after that put it back to the queue to unblock
		c.Workqueue.AddAfter(user.Name, c.authenticationOptions.AuthenticateRateLimiterDuration)
		return c.ksClient.IamV1alpha2().Users().Update(context.Background(), expect, metav1.UpdateOptions{})
	}

	return user, nil
}

func encrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// isEncrypted returns whether the given password is encrypted
func isEncrypted(password string) bool {
	// bcrypt.Cost returns the hashing cost used to create the given hashed
	cost, _ := bcrypt.Cost([]byte(password))
	// cost > 0 means the password has been encrypted
	return cost > 0
}

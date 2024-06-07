/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package im

import (
	"context"
	"fmt"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/auth"
	resourcev1beta1 "kubesphere.io/kubesphere/pkg/models/resources/v1beta1"
)

type IdentityManagementInterface interface {
	CreateUser(user *iamv1beta1.User) (*iamv1beta1.User, error)
	ListUsers(query *query.Query) (*api.ListResult, error)
	DeleteUser(username string) error
	UpdateUser(user *iamv1beta1.User) (*iamv1beta1.User, error)
	DescribeUser(username string) (*iamv1beta1.User, error)
	ModifyPassword(username string, password string) error
	ListLoginRecords(username string, query *query.Query) (*api.ListResult, error)
	PasswordVerify(username string, password string) error
}

func NewOperator(client runtimeclient.Client, resourceManager resourcev1beta1.ResourceManager, options *authentication.Options) IdentityManagementInterface {
	im := &imOperator{
		client:          client,
		options:         options,
		resourceManager: resourceManager,
	}
	return im
}

type imOperator struct {
	client          runtimeclient.Client
	resourceManager resourcev1beta1.ResourceManager
	options         *authentication.Options
}

// UpdateUser returns user information after update.
func (im *imOperator) UpdateUser(new *iamv1beta1.User) (*iamv1beta1.User, error) {
	old, err := im.fetch(new.Name)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	// keep encrypted password and user status
	new.Spec.EncryptedPassword = old.Spec.EncryptedPassword
	status := old.Status
	// only support enable or disable
	if new.Status.State == iamv1beta1.UserDisabled || new.Status.State == iamv1beta1.UserActive {
		status.State = new.Status.State
		status.LastTransitionTime = &metav1.Time{Time: time.Now()}
	}
	new.Status = status
	if err := im.client.Update(context.Background(), new); err != nil {
		return nil, err
	}
	new = new.DeepCopy()
	new.Spec.EncryptedPassword = ""
	return new, nil
}

func (im *imOperator) fetch(username string) (*iamv1beta1.User, error) {
	user := &iamv1beta1.User{}
	if err := im.client.Get(context.Background(), types.NamespacedName{Name: username}, user); err != nil {
		return nil, err
	}
	return user.DeepCopy(), nil
}

func (im *imOperator) ModifyPassword(username string, password string) error {
	user, err := im.fetch(username)
	if err != nil {
		return err
	}
	user.Spec.EncryptedPassword = password
	if err := im.client.Update(context.Background(), user); err != nil {
		return err
	}
	return nil
}

func (im *imOperator) ListUsers(query *query.Query) (*api.ListResult, error) {
	result, err := im.resourceManager.ListResources(context.Background(), iamv1beta1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcesPluralUser), "", query)
	if err != nil {
		return nil, err
	}
	items := make([]runtime.Object, 0)
	userList := result.(*iamv1beta1.UserList)
	for _, item := range userList.Items {
		out := item.DeepCopy()
		out.Spec.EncryptedPassword = ""
		items = append(items, out)
	}
	total, err := strconv.ParseInt(userList.GetContinue(), 10, 64)
	if err != nil {
		return nil, err
	}
	return &api.ListResult{Items: items, TotalItems: int(total)}, nil
}

func (im *imOperator) PasswordVerify(username string, password string) error {
	user, err := im.fetch(username)
	if err != nil {
		return err
	}
	if err = auth.PasswordVerify(user.Spec.EncryptedPassword, password); err != nil {
		return err
	}
	return nil
}

func (im *imOperator) DescribeUser(username string) (*iamv1beta1.User, error) {
	user, err := im.fetch(username)
	if err != nil {
		return nil, err
	}
	out := user.DeepCopy()
	out.Spec.EncryptedPassword = ""
	return out, nil
}

func (im *imOperator) DeleteUser(username string) error {
	user, err := im.fetch(username)
	if err != nil {
		return err
	}
	return im.client.Delete(context.Background(), user, &runtimeclient.DeleteOptions{GracePeriodSeconds: ptr.To[int64](0)})
}

func (im *imOperator) CreateUser(user *iamv1beta1.User) (*iamv1beta1.User, error) {
	if err := im.client.Create(context.Background(), user); err != nil {
		return nil, err
	}
	return user, nil
}

func (im *imOperator) ListLoginRecords(username string, q *query.Query) (*api.ListResult, error) {
	q.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", iamv1beta1.UserReferenceLabel, username))
	result, err := im.resourceManager.ListResources(context.Background(), iamv1beta1.SchemeGroupVersion.WithResource(iamv1beta1.ResourcesPluralLoginRecord), "", q)
	if err != nil {
		return nil, err
	}
	items := make([]runtime.Object, 0)
	userList := result.(*iamv1beta1.LoginRecordList)
	for _, item := range userList.Items {
		items = append(items, item.DeepCopy())
	}
	total, err := strconv.ParseInt(userList.GetContinue(), 10, 64)
	if err != nil {
		return nil, err
	}
	return &api.ListResult{Items: items, TotalItems: int(total)}, nil
}

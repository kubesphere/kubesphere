package devopsrbac

import (
	"fmt"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
	"strings"
)

func GetJenkinsRolePrefix(object metav1.Object) (string, error) {
	var workspace string
	var devops string
	var ok bool
	if workspace, ok = object.GetLabels()[constants.WorkspaceLabelKey]; !ok {
		err := fmt.Errorf("%s should have workpace label", object.GetSelfLink())
		klog.Error(err)
		return "", err
	}

	if devops, ok = object.GetLabels()[constants.DevOpsProjectLabelKey]; !ok {
		err := fmt.Errorf("%s should have devopsproject label", object.GetSelfLink())
		klog.Error(err)
		return "", err
	}
	return fmt.Sprintf("%s-%s", workspace, devops), nil
}

// getRoleTypeByName get role's type by DevOpsProjectRoleName
// valid name should be workspace-devopsproject-developer/workspace-devopsproject-owner……
func GetRoleTypeByRoleName(name string) (string, error) {
	parts := strings.Split(name, "-")
	if len(parts) < 3 {
		err := fmt.Errorf("%s is not a valid role name", name)
		klog.Error(err)
		return "", err
	}
	if !reflectutils.In(parts[len(parts)-1], devops.AllRoleSlice) {
		err := fmt.Errorf("err role [%s] not in [%s]", name,
			devops.AllRoleSlice)
		klog.Errorf("%+v", err)
	}
	return parts[len(parts)-1], nil
}

func RBACSubjectsToStringSlice(subjects []rbacv1.Subject) []string {
	var result []string
	if len(subjects) == 0 {
		return make([]string, 0)
	}
	for _, subject := range subjects {
		result = append(result, subject.Name)
	}
	return result

}

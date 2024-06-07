/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package serviceaccount

import (
	"strings"

	"k8s.io/apiserver/pkg/authentication/user"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
)

func IsServiceAccountToken(subjectName string) bool {
	if !strings.HasPrefix(subjectName, corev1alpha1.ServiceAccountTokenPrefix) {
		return false
	}
	split := strings.Split(subjectName, ":")

	return len(split) == 4
}

func GetSecretName(info user.Info) (name, namespace string) {
	extra := info.GetExtra()
	if value, ok := extra[corev1alpha1.ServiceAccountTokenExtraSecretName]; ok {
		name = value[0]
	}

	if value, ok := extra[corev1alpha1.ServiceAccountTokenExtraSecretNamespace]; ok {
		namespace = value[0]
	}
	return
}

func SplitUsername(username string) (name, namespace string) {
	if !strings.HasPrefix(username, corev1alpha1.ServiceAccountTokenPrefix) {
		return "", ""
	}
	split := strings.Split(username, ":")
	if len(split) != 4 {
		return "", ""
	}
	return split[3], split[2]
}

// MatchesUsername checks whether the provided username matches the namespace and name without
// allocating. Use this when checking a service account namespace and name against a known string.
func MatchesUsername(namespace, name string, username string) bool {
	if !strings.HasPrefix(username, corev1alpha1.ServiceAccountTokenPrefix) {
		return false
	}
	username = username[len(corev1alpha1.ServiceAccountTokenPrefix):]

	if !strings.HasPrefix(username, namespace) {
		return false
	}
	username = username[len(namespace):]

	if !strings.HasPrefix(username, ":") {
		return false
	}
	username = username[len(":"):]

	return username == name
}

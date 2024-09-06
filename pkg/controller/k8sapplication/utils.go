/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package k8sapplication

import (
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AppLabel                = "app"
	VersionLabel            = "version"
	ApplicationNameLabel    = "app.kubernetes.io/name"
	ApplicationVersionLabel = "app.kubernetes.io/version"
)

// resource with these following labels considered as part of servicemesh
var ApplicationLabels = [...]string{
	ApplicationNameLabel,
	ApplicationVersionLabel,
	AppLabel,
}

// resource with these following labels considered as part of kubernetes-sigs/application
var AppLabels = [...]string{
	ApplicationNameLabel,
	ApplicationVersionLabel,
}

var TrimChars = [...]string{".", "_", "-"}

// normalize version names
// strip [_.-]
func NormalizeVersionName(version string) string {
	for _, char := range TrimChars {
		version = strings.ReplaceAll(version, char, "")
	}
	return version
}

func GetApplictionName(lbs map[string]string) string {
	if name, ok := lbs[ApplicationNameLabel]; ok {
		return name
	}
	return ""

}

func GetComponentName(meta *v1.ObjectMeta) string {
	if len(meta.Labels[AppLabel]) > 0 {
		return meta.Labels[AppLabel]
	}
	return ""
}

func GetComponentVersion(meta *v1.ObjectMeta) string {
	if len(meta.Labels[VersionLabel]) > 0 {
		return meta.Labels[VersionLabel]
	}
	return ""
}

func ExtractApplicationLabels(meta *v1.ObjectMeta) map[string]string {

	labels := make(map[string]string, len(ApplicationLabels))
	for _, label := range ApplicationLabels {
		if _, ok := meta.Labels[label]; !ok {
			return nil
		} else {
			labels[label] = meta.Labels[label]
		}
	}

	return labels
}

func IsApplicationComponent(lbs map[string]string) bool {

	for _, label := range ApplicationLabels {
		if _, ok := lbs[label]; !ok {
			return false
		}
	}

	return true
}

// Whether it belongs to kubernetes-sigs/application or not
func IsAppComponent(lbs map[string]string) bool {

	for _, label := range AppLabels {
		if _, ok := lbs[label]; !ok {
			return false
		}
	}

	return true
}

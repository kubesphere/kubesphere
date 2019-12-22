package servicemesh

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AppLabel                = "app"
	VersionLabel            = "version"
	ApplicationNameLabel    = "app.kubernetes.io/name"
	ApplicationVersionLabel = "app.kubernetes.io/version"
)

var ApplicationLabels = [...]string{
	ApplicationNameLabel,
	ApplicationVersionLabel,
	AppLabel,
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

func GetComponentName(meta *metav1.ObjectMeta) string {
	if len(meta.Labels[AppLabel]) > 0 {
		return meta.Labels[AppLabel]
	}
	return ""
}

func GetComponentVersion(meta *metav1.ObjectMeta) string {
	if len(meta.Labels[VersionLabel]) > 0 {
		return meta.Labels[VersionLabel]
	}
	return ""
}

func ExtractApplicationLabels(meta *metav1.ObjectMeta) map[string]string {

	labels := make(map[string]string, 0)
	for _, label := range ApplicationLabels {
		if len(meta.Labels[label]) == 0 {
			return nil
		} else {
			labels[label] = meta.Labels[label]
		}
	}

	return labels
}

func IsApplicationComponent(meta *metav1.ObjectMeta) bool {

	for _, label := range ApplicationLabels {
		if len(meta.Labels[label]) == 0 {
			return false
		}
	}

	return true
}

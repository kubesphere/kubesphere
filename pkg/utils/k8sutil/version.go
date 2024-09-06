/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package k8sutil

import (
	"github.com/Masterminds/semver/v3"
)

func ServeBatchV1beta1(k8sVersion *semver.Version) bool {
	c, _ := semver.NewConstraint("< 1.21")
	return c.Check(k8sVersion)
}

func ServeAutoscalingV2beta2(k8sVersion *semver.Version) bool {
	c, _ := semver.NewConstraint("< 1.23")
	return c.Check(k8sVersion)
}

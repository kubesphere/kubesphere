/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package k8sutil

import (
	"github.com/Masterminds/semver/v3"
)

func ServeBatchV1beta1(k8sVersion *semver.Version) bool {
	// add "-0" to make the prerelease version compatible.
	c, _ := semver.NewConstraint("< 1.21.0-0")
	return c.Check(k8sVersion)
}

func ServeAutoscalingV2beta2(k8sVersion *semver.Version) bool {
	// add "-0" to make the prerelease version compatible.
	c, _ := semver.NewConstraint("< 1.23.0-0")
	return c.Check(k8sVersion)
}

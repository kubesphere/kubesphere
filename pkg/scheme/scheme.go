/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package scheme

import (
	kubespherescheme "kubesphere.io/client-go/kubesphere/scheme"
)

// Scheme contains all types of custom Scheme and kubernetes client-go Scheme.
var Scheme = kubespherescheme.Scheme
var Codecs = kubespherescheme.Codecs

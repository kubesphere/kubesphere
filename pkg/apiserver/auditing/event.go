/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package auditing

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/apis/audit"
)

type Event struct {
	// The workspace which this audit event happened
	Workspace string
	// The cluster which this audit event happened
	Cluster string
	// Message send to user.
	Message  string
	HostName string
	HostIP   string

	audit.Event
}

type Object struct {
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

/*


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

package v2beta2

import (
	"time"

	"github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SilenceSpec defines the desired state of Silence
type SilenceSpec struct {
	// whether the silence is enabled
	Enabled *bool                 `json:"enabled,omitempty"`
	Matcher *metav1.LabelSelector `json:"matcher"`
	// The start time during which the silence is active.
	//
	// +kubebuilder:validation:Format: date-time
	StartsAt *metav1.Time `json:"startsAt,omitempty"`
	// The schedule in Cron format.
	// If set the silence will be active periodicity, and the startsAt will be invalid.
	Schedule string `json:"schedule,omitempty"`
	// The time range during which the silence is active.
	// If not set, the silence will be active ever.
	Duration *metav1.Duration `json:"duration,omitempty"`
}

// SilenceStatus defines the observed state of Silence
type SilenceStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,categories=notification-manager
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +genclient
// +genclient:nonNamespaced

// Silence is the Schema for the Silence API
type Silence struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SilenceSpec   `json:"spec,omitempty"`
	Status SilenceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SilenceList contains a list of Silence
type SilenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Silence `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Silence{}, &SilenceList{})
}

func (s *Silence) IsActive() bool {

	if s.Spec.Enabled != nil && !*s.Spec.Enabled {
		return false
	}

	if s.Spec.Schedule != "" {

		if s.Spec.Duration == nil {
			return true
		}

		schedule, _ := cron.ParseStandard(s.Spec.Schedule)
		if schedule.Next(time.Now()) == schedule.Next(time.Now().Add(-(*s.Spec.Duration).Duration)) {
			return false
		} else {
			return true
		}
	} else if s.Spec.StartsAt != nil {
		if s.Spec.StartsAt.After(time.Now()) {
			return false
		}

		if s.Spec.Duration == nil {
			return true
		}

		if s.Spec.StartsAt.Add((*s.Spec.Duration).Duration).After(time.Now()) {
			return true
		}

		return false
	} else {
		return true
	}
}

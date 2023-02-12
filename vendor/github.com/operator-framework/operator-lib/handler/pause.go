// Copyright 2021 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"github.com/operator-framework/operator-lib/internal/annotation"

	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// NewPause returns an event handler that filters out objects with a truthy "paused" annotation.
// When an annotation with key string key is present on an object and has a truthy value, ex. "true",
// the watch constructed with this event handler will not add events for that object to the queue.
// Key string key must be a valid annotation key.
//
// A note on security: since users that can CRUD a particular API can apply or remove annotations with
// default cluster admission controllers, this same set of users can therefore start or stop reconciliation
// of objects via this pause mechanism. If this is a concern, configure an admission webhook to enforce
// a stricter annotation modification policy. See AdmissionReview configuration for user info available
// to a webhook:
// https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#request
func NewPause(key string) (handler.EventHandler, error) {
	return annotation.NewFalsyEventHandler(key, annotation.Options{Log: log})
}

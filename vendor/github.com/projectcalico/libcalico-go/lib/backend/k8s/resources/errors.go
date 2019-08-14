// Copyright (c) 2016 Tigera, Inc. All rights reserved.

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

package resources

import (
	"strings"

	"github.com/projectcalico/libcalico-go/lib/errors"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

// K8sErrorToCalico returns the equivalent libcalico error for the given
// kubernetes error.
func K8sErrorToCalico(ke error, id interface{}) error {
	if ke == nil {
		return nil
	}

	if kerrors.IsAlreadyExists(ke) {
		return errors.ErrorResourceAlreadyExists{
			Err:        ke,
			Identifier: id,
		}
	}
	if kerrors.IsNotFound(ke) {
		return errors.ErrorResourceDoesNotExist{
			Err:        ke,
			Identifier: id,
		}
	}
	if kerrors.IsForbidden(ke) || kerrors.IsUnauthorized(ke) {
		return errors.ErrorConnectionUnauthorized{
			Err: ke,
		}
	}
	if kerrors.IsConflict(ke) {
		// Treat precondition errors as not found.
		if strings.Contains(ke.Error(), "UID in precondition") {
			return errors.ErrorResourceDoesNotExist{
				Err:        ke,
				Identifier: id,
			}
		}
		return errors.ErrorResourceUpdateConflict{
			Err:        ke,
			Identifier: id,
		}
	}
	return errors.ErrorDatastoreError{
		Err:        ke,
		Identifier: id,
	}
}

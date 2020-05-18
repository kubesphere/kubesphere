/*
 *
 * Copyright 2020 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package token

import "time"

// Issuer issues token to user, tokens are required to perform mutating requests to resources
type Issuer interface {
	// IssueTo issues a token a User, return error if issuing process failed
	IssueTo(user User, expiresIn time.Duration) (string, error)

	// Verify verifies a token, and return a User if it's a valid token, otherwise return error
	Verify(string) (User, error)

	// Revoke a token,
	Revoke(token string) error
}

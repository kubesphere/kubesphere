/*
Copyright 2021 The KubeSphere Authors.

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

package oauth

import "fmt"

// The following error type is defined in https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
var (
	// ErrorInvalidClient
	// Client authentication failed (e.g., unknown client, no
	// client authentication included, or unsupported
	// authentication method).  The authorization server MAY
	// return an HTTP 401 (Unauthorized) status code to indicate
	// which HTTP authentication schemes are supported.  If the
	// client attempted to authenticate via the "Authorization"
	// request header field, the authorization server MUST
	// respond with an HTTP 401 (Unauthorized) status code and
	// include the "WWW-Authenticate" response header field
	// matching the authentication scheme used by the client.
	ErrorInvalidClient = Error{Type: "invalid_client"}

	// ErrorInvalidRequest The request is missing a required parameter,
	// includes an unsupported parameter value (other than grant type),
	// repeats a parameter, includes multiple credentials,
	// utilizes more than one mechanism for authenticating the client,
	// or is otherwise malformed.
	ErrorInvalidRequest = Error{Type: "invalid_request"}

	// ErrorInvalidGrant
	// The provided authorization grant (e.g., authorization code,
	// resource owner credentials) or refresh token is invalid, expired, revoked,
	// does not match the redirection URI used in the authorization request,
	// or was issued to another client.
	ErrorInvalidGrant = Error{Type: "invalid_grant"}

	// ErrorUnsupportedGrantType
	// The authorization grant type is not supported by the authorization server.
	ErrorUnsupportedGrantType = Error{Type: "unsupported_grant_type"}

	ErrorUnsupportedResponseType = Error{Type: "unsupported_response_type"}

	// ErrorUnauthorizedClient
	// The authenticated client is not authorized to use this authorization grant type.
	ErrorUnauthorizedClient = Error{Type: "unauthorized_client"}

	// ErrorInvalidScope The requested scope is invalid, unknown, malformed,
	// or exceeds the scope granted by the resource owner.
	ErrorInvalidScope = Error{Type: "invalid_scope"}

	// ErrorLoginRequired The Authorization Server requires End-User authentication.
	// This error MAY be returned when the prompt parameter value in the Authentication Request is none,
	// but the Authentication Request cannot be completed without displaying a user interface
	// for End-User authentication.
	ErrorLoginRequired = Error{Type: "login_required"}

	// ErrorServerError
	// The authorization server encountered an unexpected
	// condition that prevented it from fulfilling the request.
	// (This error code is needed because a 500 Internal Server
	// Error HTTP status code cannot be returned to the client
	// via an HTTP redirect.)
	ErrorServerError = Error{Type: "server_error"}
)

func NewInvalidRequest(error error) Error {
	err := ErrorInvalidRequest
	err.Description = error.Error()
	return err
}

func NewInvalidScope(error error) Error {
	err := ErrorInvalidScope
	err.Description = error.Error()
	return err
}

func NewInvalidClient(error error) Error {
	err := ErrorInvalidClient
	err.Description = error.Error()
	return err
}

func NewInvalidGrant(error error) Error {
	err := ErrorInvalidGrant
	err.Description = error.Error()
	return err
}

func NewServerError(error error) Error {
	err := ErrorServerError
	err.Description = error.Error()
	return err
}

// Error wrapped OAuth error Response, for more details: https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
// The authorization server responds with an HTTP 400 (Bad Request)
// status code (unless specified otherwise) and includes the following
// parameters with the response:
type Error struct {
	// Type REQUIRED
	// A single ASCII [USASCII] error code from the following:
	// Values for the "error" parameter MUST NOT include characters
	// outside the set %x20-21 / %x23-5B / %x5D-7E.
	Type string `json:"error"`
	// Description OPTIONAL.  Human-readable ASCII [USASCII] text providing
	// additional information, used to assist the client developer in
	// understanding the error that occurred.
	// Values for the "error_description" parameter MUST NOT include
	// characters outside the set %x20-21 / %x23-5B / %x5D-7E.
	Description string `json:"error_description,omitempty"`
}

func (e Error) Error() string {
	return fmt.Sprintf("error=\"%s\", error_description=\"%s\"", e.Type, e.Description)
}

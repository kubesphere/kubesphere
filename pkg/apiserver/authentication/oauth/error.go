/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package oauth

import "fmt"

type ErrorType string

// The following error type is defined in https://datatracker.ietf.org/doc/html/rfc6749#section-4.1.2.1
const (
	// InvalidClient
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
	InvalidClient ErrorType = "invalid_client"

	// InvalidRequest
	// The request is missing a required parameter, includes an unsupported parameter value (other than grant type),
	// repeats a parameter, includes multiple credentials, utilizes more than one mechanism for authenticating the client,
	// or is otherwise malformed.
	InvalidRequest ErrorType = "invalid_request"

	// InvalidGrant
	// The provided authorization grant (e.g., authorization code,
	// resource owner credentials) or refresh token is invalid, expired, revoked,
	// does not match the redirection URI used in the authorization request,
	// or was issued to another client.
	InvalidGrant ErrorType = "invalid_grant"

	// UnsupportedGrantType
	// The authorization grant type is not supported by the authorization server.
	UnsupportedGrantType ErrorType = "unsupported_grant_type"

	// UnsupportedResponseType
	// The authorization server does not support obtaining an authorization code using this method.
	UnsupportedResponseType ErrorType = "unsupported_response_type"

	// UnauthorizedClient
	// The authenticated client is not authorized to use this authorization grant type.
	UnauthorizedClient ErrorType = "unauthorized_client"

	// InvalidScope The requested scope is invalid, unknown, malformed,
	// or exceeds the scope granted by the resource owner.
	InvalidScope ErrorType = "invalid_scope"

	// LoginRequired The Authorization Server requires End-User authentication.
	// This error MAY be returned when the prompt parameter value in the Authentication Request is none,
	// but the Authentication Request cannot be completed without displaying a user interface
	// for End-User authentication.
	LoginRequired ErrorType = "login_required"

	// InteractionRequired
	// The Authorization Server requires End-User interaction of some form to proceed.
	// This error MAY be returned when the prompt parameter value in the Authentication Request is none,
	// but the Authentication Request cannot be completed without displaying a user interface for End-User interaction.
	InteractionRequired ErrorType = "interaction_required"

	// ServerError
	// The authorization server encountered an unexpected
	// condition that prevented it from fulfilling the request.
	// (This error code is needed because a 500 Internal Server
	// Error HTTP status code cannot be returned to the client
	// via an HTTP redirect.)
	ServerError ErrorType = "server_error"
)

func NewError(errorType ErrorType, description string) *Error {
	return &Error{
		Type:        errorType,
		Description: description,
	}
}

func NewInvalidRequest(description string) *Error {
	return &Error{
		Type:        InvalidRequest,
		Description: description,
	}
}

func NewInvalidScope(description string) *Error {
	return &Error{
		Type:        InvalidScope,
		Description: description,
	}
}

func NewInvalidClient(description string) *Error {
	return &Error{
		Type:        InvalidClient,
		Description: description,
	}
}

func NewInvalidGrant(description string) *Error {
	return &Error{
		Type:        InvalidGrant,
		Description: description,
	}
}

func NewServerError(description string) *Error {
	return &Error{
		Type:        ServerError,
		Description: description,
	}
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
	Type ErrorType `json:"error"`
	// Description OPTIONAL.  Human-readable ASCII [USASCII] text providing
	// additional information, used to assist the client developer in
	// understanding the error that occurred.
	// Values for the "error_description" parameter MUST NOT include
	// characters outside the set %x20-21 / %x23-5B / %x5D-7E.
	Description string `json:"error_description,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("error=\"%s\", error_description=\"%s\"", e.Type, e.Description)
}

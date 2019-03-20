/*
Copyright 2018 The Knative Authors

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

package v1alpha1

import (
	"github.com/knative/pkg/apis/istio/common/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VirtualService
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PolicySpec `json:"spec"`
}

// Policy defines what authentication methods can be accepted on workload(s),
// and if authenticated, which method/certificate will set the request principal
// (i.e request.auth.principal attribute).
//
// Authentication policy is composed of 2-part authentication:
// - peer: verify caller service credentials. This part will set source.user
// (peer identity).
// - origin: verify the origin credentials. This part will set request.auth.user
// (origin identity), as well as other attributes like request.auth.presenter,
// request.auth.audiences and raw claims. Note that the identity could be
// end-user, service account, device etc.
//
// Last but not least, the principal binding rule defines which identity (peer
// or origin) should be used as principal. By default, it uses peer.
//
// Examples:
//
// Policy to enable mTLS for all services in namespace frod
//
// apiVersion: authentication.istio.io/v1alpha1
// kind: Policy
// metadata:
//   name: mTLS_enable
//   namespace: frod
// spec:
//   peers:
//   - mtls:
//
// Policy to disable mTLS for "productpage" service
//
// apiVersion: authentication.istio.io/v1alpha1
// kind: Policy
// metadata:
//   name: mTLS_disable
//   namespace: frod
// spec:
//   targets:
//   - name: productpage
//
// Policy to require mTLS for peer authentication, and JWT for origin authenticationn
// for productpage:9000. Principal is set from origin identity.
//
// apiVersion: authentication.istio.io/v1alpha1
// kind: Policy
// metadata:
//   name: mTLS_enable
//   namespace: frod
// spec:
//   target:
//   - name: productpage
//     ports:
//     - number: 9000
//   peers:
//   - mtls:
//   origins:
//   - jwt:
//       issuer: "https://securetoken.google.com"
//       audiences:
//       - "productpage"
//       jwksUri: "https://www.googleapis.com/oauth2/v1/certs"
//       jwt_headers:
//       - "x-goog-iap-jwt-assertion"
//   principaBinding: USE_ORIGIN
//
// Policy to require mTLS for peer authentication, and JWT for origin authenticationn
// for productpage:9000, but allow origin authentication failed. Principal is set
// from origin identity.
// Note: this example can be used for use cases when we want to allow request from
// certain peers, given it comes with an approperiate authorization poicy to check
// and reject request accoridingly.
//
// apiVersion: authentication.istio.io/v1alpha1
// kind: Policy
// metadata:
//   name: mTLS_enable
//   namespace: frod
// spec:
//   target:
//   - name: productpage
//     ports:
//     - number: 9000
//   peers:
//   - mtls:
//   origins:
//   - jwt:
//       issuer: "https://securetoken.google.com"
//       audiences:
//       - "productpage"
//       jwksUri: "https://www.googleapis.com/oauth2/v1/certs"
//       jwt_headers:
//       - "x-goog-iap-jwt-assertion"
//   originIsOptional: true
//   principalBinding: USE_ORIGIN
type PolicySpec struct {
	// List rules to select destinations that the policy should be applied on.
	// If empty, policy will be used on all destinations in the same namespace.
	Targets []TargetSelector `json:"targets,omitempty"`

	// List of authentication methods that can be used for peer authentication.
	// They will be evaluated in order; the first validate one will be used to
	// set peer identity (source.user) and other peer attributes. If none of
	// these methods pass, and peer_is_optional flag is false (see below),
	// request will be rejected with authentication failed error (401).
	// Leave the list empty if peer authentication is not required
	Peers []PeerAuthenticationMethod `json:"peers,omitempty"`

	// Set this flag to true to accept request (for peer authentication perspective),
	// even when none of the peer authentication methods defined above satisfied.
	// Typically, this is used to delay the rejection decision to next layer (e.g
	// authorization).
	// This flag is ignored if no authentication defined for peer (peers field is empty).
	PeerIsOptional bool `json:"peerIsOptional,omitempty"`

	// List of authentication methods that can be used for origin authentication.
	// Similar to peers, these will be evaluated in order; the first validate one
	// will be used to set origin identity and attributes (i.e request.auth.user,
	// request.auth.issuer etc). If none of these methods pass, and origin_is_optional
	// is false (see below), request will be rejected with authentication failed
	// error (401).
	// Leave the list empty if origin authentication is not required.
	Origins []OriginAuthenticationMethod `json:"origins,omitempty"`

	// Set this flag to true to accept request (for origin authentication perspective),
	// even when none of the origin authentication methods defined above satisfied.
	// Typically, this is used to delay the rejection decision to next layer (e.g
	// authorization).
	// This flag is ignored if no authentication defined for origin (origins field is empty).
	OriginIsOptional bool `json:"originIsOptional,omitempty"`

	// Define whether peer or origin identity should be use for principal. Default
	// value is USE_PEER.
	// If peer (or origin) identity is not available, either because of peer/origin
	// authentication is not defined, or failed, principal will be left unset.
	// In other words, binding rule does not affect the decision to accept or
	// reject request.
	PrincipalBinding PrincipalBinding `json:"principalBinding,omitempty"`
}

// TargetSelector defines a matching rule to a service/destination.
type TargetSelector struct {
	// REQUIRED. The name must be a short name from the service registry. The
	// fully qualified domain name will be resolved in a platform specific manner.
	Name string `json:"name"`

	// Specifies the ports on the destination. Leave empty to match all ports
	// that are exposed.
	Ports []PortSelector `json:"ports,omitempty"`
}

// PortSelector specifies the name or number of a port to be used for
// matching targets for authenticationn policy. This is copied from
// networking API to avoid dependency.
type PortSelector struct {
	// It is required to specify exactly one of the fields:
	// Number or Name

	// Valid port number
	Number uint32 `json:"number,omitempty"`

	// Port name
	Name string `json:"name,omitempty"`
}

// PeerAuthenticationMethod defines one particular type of authentication, e.g
// mutual TLS, JWT etc, (no authentication is one type by itself) that can
// be used for peer authentication.
// The type can be progammatically determine by checking the type of the
// "params" field.
type PeerAuthenticationMethod struct {
	// It is required to specify exactly one of the fields:
	// Mtls or Jwt
	// Set if mTLS is used.
	Mtls *MutualTls `json:"mtls,omitempty"`

	// Set if JWT is used. This option is not yet available.
	Jwt *Jwt `json:"jwt,omitempty"`
}

// Defines the acceptable connection TLS mode.
type Mode string

const (
	// Client cert must be presented, connection is in TLS.
	ModeStrict Mode = "STRICT"

	// Connection can be either plaintext or TLS, and client cert can be omitted.
	ModePermissive Mode = "PERMISSIVE"
)

// TLS authentication params.
type MutualTls struct {

	// WILL BE DEPRECATED, if set, will translates to `TLS_PERMISSIVE` mode.
	// Set this flag to true to allow regular TLS (i.e without client x509
	// certificate). If request carries client certificate, identity will be
	// extracted and used (set to peer identity). Otherwise, peer identity will
	// be left unset.
	// When the flag is false (default), request must have client certificate.
	AllowTls bool `json:"allowTls,omitempty"`

	// Defines the mode of mTLS authentication.
	Mode Mode `json:"mode,omitempty"`
}

// JSON Web Token (JWT) token format for authentication as defined by
// https://tools.ietf.org/html/rfc7519. See [OAuth
// 2.0](https://tools.ietf.org/html/rfc6749) and [OIDC
// 1.0](http://openid.net/connect) for how this is used in the whole
// authentication flow.
//
// Example,
//
// issuer: https://example.com
// audiences:
// - bookstore_android.apps.googleusercontent.com
//   bookstore_web.apps.googleusercontent.com
// jwksUri: https://example.com/.well-known/jwks.json
//
type Jwt struct {
	// Identifies the issuer that issued the JWT. See
	// [issuer](https://tools.ietf.org/html/rfc7519#section-4.1.1)
	// Usually a URL or an email address.
	//
	// Example: https://securetoken.google.com
	// Example: 1234567-compute@developer.gserviceaccount.com
	Issuer string `json:"issuer,omitempty"`

	// The list of JWT
	// [audiences](https://tools.ietf.org/html/rfc7519#section-4.1.3).
	// that are allowed to access. A JWT containing any of these
	// audiences will be accepted.
	//
	// The service name will be accepted if audiences is empty.
	//
	// Example:
	//
	// ```yaml
	// audiences:
	// - bookstore_android.apps.googleusercontent.com
	//   bookstore_web.apps.googleusercontent.com
	// ```
	Audiences []string `json:"audiences,omitempty"`

	// URL of the provider's public key set to validate signature of the
	// JWT. See [OpenID
	// Discovery](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata).
	//
	// Optional if the key set document can either (a) be retrieved from
	// [OpenID
	// Discovery](https://openid.net/specs/openid-connect-discovery-1_0.html) of
	// the issuer or (b) inferred from the email domain of the issuer (e.g. a
	// Google service account).
	//
	// Example: https://www.googleapis.com/oauth2/v1/certs
	JwksUri string `json:"jwksUri,omitempty"`

	// Two fields below define where to extract the JWT from an HTTP request.
	//
	// If no explicit location is specified the following default
	// locations are tried in order:
	//
	//     1) The Authorization header using the Bearer schema,
	//        e.g. Authorization: Bearer <token>. (see
	//        [Authorization Request Header
	//        Field](https://tools.ietf.org/html/rfc6750#section-2.1))
	//
	//     2) `access_token` query parameter (see
	//     [URI Query Parameter](https://tools.ietf.org/html/rfc6750#section-2.3))
	// JWT is sent in a request header. `header` represents the
	// header name.
	//
	// For example, if `header=x-goog-iap-jwt-assertion`, the header
	// format will be x-goog-iap-jwt-assertion: <JWT>.
	JwtHeaders []string `json:"jwtHeaders,omitempty"`

	// JWT is sent in a query parameter. `query` represents the
	// query parameter name.
	//
	// For example, `query=jwt_token`.
	JwtParams []string `json:"jwtParams,omitempty"`

	// URL paths that should be excluded from the JWT validation. If the request path is matched,
	// the JWT validation will be skipped and the request will proceed regardless.
	// This is useful to keep a couple of URLs public for external health checks.
	// Example: "/health_check", "/status/cpu_usage".
	ExcludedPaths []v1alpha1.StringMatch `json:"excludedPaths,omitempty"`
}

// OriginAuthenticationMethod defines authentication method/params for origin
// authentication. Origin could be end-user, device, delegate service etc.
// Currently, only JWT is supported for origin authentication.
type OriginAuthenticationMethod struct {
	// Jwt params for the method.
	Jwt *Jwt `json:"jwt,omitempty"`
}

// Associates authentication with request principal.
type PrincipalBinding string

const (
	// Principal will be set to the identity from peer authentication.
	PrincipalBindingUserPeer PrincipalBinding = "USE_PEER"
	// Principal will be set to the identity from peer authentication.
	PrincipalBindingUserOrigin PrincipalBinding = "USE_ORIGIN"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// PolicyLIst is a list of Policy resources
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Policy `json:"items"`
}

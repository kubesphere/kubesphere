package cas

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
)

// AuthenticationError Code values
const (
	INVALID_REQUEST            = "INVALID_REQUEST"
	INVALID_TICKET_SPEC        = "INVALID_TICKET_SPEC"
	UNAUTHORIZED_SERVICE       = "UNAUTHORIZED_SERVICE"
	UNAUTHORIZED_SERVICE_PROXY = "UNAUTHORIZED_SERVICE_PROXY"
	INVALID_PROXY_CALLBACK     = "INVALID_PROXY_CALLBACK"
	INVALID_TICKET             = "INVALID_TICKET"
	INVALID_SERVICE            = "INVALID_SERVICE"
	INTERNAL_ERROR             = "INTERNAL_ERROR"
)

// AuthenticationError represents a CAS AuthenticationFailure response
type AuthenticationError struct {
	Code    string
	Message string
}

// AuthenticationError provides a differentiator for casting.
func (e AuthenticationError) AuthenticationError() bool {
	return true
}

// Error returns the AuthenticationError as a string
func (e AuthenticationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// AuthenticationResponse captures authenticated user information
type AuthenticationResponse struct {
	User                string         // Users login name
	ProxyGrantingTicket string         // Proxy Granting Ticket
	Proxies             []string       // List of proxies
	AuthenticationDate  time.Time      // Time at which authentication was performed
	IsNewLogin          bool           // Whether new authentication was used to grant the service ticket
	IsRememberedLogin   bool           // Whether a long term token was used to grant the service ticket
	MemberOf            []string       // List of groups which the user is a member of
	Attributes          UserAttributes // Additional information about the user
}

// UserAttributes represents additional data about the user
type UserAttributes map[string][]string

// Get retrieves an attribute by name.
//
// Attributes are stored in arrays. Get will only return the first element.
func (a UserAttributes) Get(name string) string {
	if v, ok := a[name]; ok {
		return v[0]
	}

	return ""
}

// Add appends a new attribute.
func (a UserAttributes) Add(name, value string) {
	a[name] = append(a[name], value)
}

// ParseServiceResponse returns a successful response or an error
func ParseServiceResponse(data []byte) (*AuthenticationResponse, error) {
	var x xmlServiceResponse

	if err := xml.Unmarshal(data, &x); err != nil {
		return nil, err
	}

	if x.Failure != nil {
		msg := strings.TrimSpace(x.Failure.Message)
		err := &AuthenticationError{Code: x.Failure.Code, Message: msg}
		return nil, err
	}

	r := &AuthenticationResponse{
		User:                x.Success.User,
		ProxyGrantingTicket: x.Success.ProxyGrantingTicket,
		Attributes:          make(UserAttributes),
	}

	if p := x.Success.Proxies; p != nil {
		r.Proxies = p.Proxies
	}

	if a := x.Success.Attributes; a != nil {
		r.AuthenticationDate = a.AuthenticationDate
		r.IsRememberedLogin = a.LongTermAuthenticationRequestTokenUsed
		r.IsNewLogin = a.IsFromNewLogin
		r.MemberOf = a.MemberOf

		if a.UserAttributes != nil {
			for _, ua := range a.UserAttributes.Attributes {
				if ua.Name == "" {
					continue
				}

				r.Attributes.Add(ua.Name, strings.TrimSpace(ua.Value))
			}

			for _, ea := range a.UserAttributes.AnyAttributes {
				r.Attributes.Add(ea.XMLName.Local, strings.TrimSpace(ea.Value))
			}
		}

		if a.ExtraAttributes != nil {
			for _, ea := range a.ExtraAttributes {
				r.Attributes.Add(ea.XMLName.Local, strings.TrimSpace(ea.Value))
			}
		}
	}

	for _, ea := range x.Success.ExtraAttributes {
		addRubycasAttribute(r.Attributes, ea.XMLName.Local, strings.TrimSpace(ea.Value))
	}

	return r, nil
}

// addRubycasAttribute handles RubyCAS style additional attributes.
func addRubycasAttribute(attributes UserAttributes, key, value string) {
	if !strings.HasPrefix(value, "---") {
		attributes.Add(key, value)
		return
	}

	if value == "--- true" {
		attributes.Add(key, "true")
		return
	}

	if value == "--- false" {
		attributes.Add(key, "false")
		return
	}

	var decoded interface{}
	if err := yaml.Unmarshal([]byte(value), &decoded); err != nil {
		attributes.Add(key, err.Error())
		return
	}

	switch reflect.TypeOf(decoded).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(decoded)

		for i := 0; i < s.Len(); i++ {
			e := s.Index(i).Interface()

			switch reflect.TypeOf(e).Kind() {
			case reflect.String:
				attributes.Add(key, e.(string))
			}
		}
	case reflect.String:
		s := reflect.ValueOf(decoded).Interface()
		attributes.Add(key, s.(string))
	default:
		if glog.V(2) {
			kind := reflect.TypeOf(decoded).Kind()
			glog.Warningf("cas: service response: unable to parse %v value: %#v (kind: %v)", key, decoded, kind)
		}
	}

	return
}

package ldap

import (
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
)

// Interface defines CRUD behaviors of manipulating users
type Interface interface {
	// Create create a new user in ldap
	Create(user *iamv1alpha2.User) error

	// Update updates a user information, return error if user not exists
	Update(user *iamv1alpha2.User) error

	// Delete deletes a user from ldap, return nil if user not exists
	Delete(name string) error

	// Get gets a user by its username from ldap, return ErrUserNotExists if user not exists
	Get(name string) (*iamv1alpha2.User, error)

	// Authenticate checks if (name, password) is valid, return ErrInvalidCredentials if not
	Authenticate(name string, password string) error
}

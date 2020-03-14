package ldap

import "kubesphere.io/kubesphere/pkg/api/iam"

// Interface defines CRUD behaviors of manipulating users
type Interface interface {
    // Create create a new user in ldap
    Create(user *iam.User) error

    // Update updates a user information, return error if user not exists
    Update(user *iam.User) error

    // Delete deletes a user from ldap, return nil if user not exists
    Delete(name string) error

    // Get gets a user by its username from ldap, return ErrUserNotExists if user not exists
    Get(name string) (*iam.User, error)

    // Verify checks if (name, password) is valid, return ErrInvalidCredentials if not
    Verify(name string, password string) error
}

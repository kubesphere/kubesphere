package ldap

import (
	"kubesphere.io/kubesphere/pkg/api/iam"
	"time"
)

// simpleLdap is a implementation of ldap.Interface, you should never use this in production env!
type simpleLdap struct {
	store map[string]*iam.User
}

func NewSimpleLdap() Interface {
	sl := &simpleLdap{
		store: map[string]*iam.User{},
	}

	// initialize with a admin user
	admin := &iam.User{
		Username:    "admin",
		Email:       "admin@kubesphere.io",
		Lang:        "eng",
		Description: "administrator",
		CreateTime:  time.Now(),
		Groups:      nil,
		Password:    "P@88w0rd",
	}
	sl.store[admin.Username] = admin
	return sl
}

func (s simpleLdap) Create(user *iam.User) error {
	s.store[user.Username] = user
	return nil
}

func (s simpleLdap) Update(user *iam.User) error {
	_, err := s.Get(user.Username)
	if err != nil {
		return err
	}
	s.store[user.Username] = user
	return nil
}

func (s simpleLdap) Delete(name string) error {
	_, err := s.Get(name)
	if err != nil {
		return err
	}
	delete(s.store, name)
	return nil
}

func (s simpleLdap) Get(name string) (*iam.User, error) {
	if user, ok := s.store[name]; !ok {
		return nil, ErrUserNotExists
	} else {
		return user, nil
	}
}

func (s simpleLdap) Verify(name string, password string) error {
	if user, err := s.Get(name); err != nil {
		return err
	} else {
		if user.Password != password {
			return ErrInvalidCredentials
		}
	}

	return nil
}

package iam

import (
	"kubesphere.io/kubesphere/pkg/server/errors"
	"time"
)

type User struct {
	Name        string    `json:"username"`
	UID         string    `json:"uid"`
	Email       string    `json:"email"`
	Lang        string    `json:"lang,omitempty"`
	Description string    `json:"description"`
	CreateTime  time.Time `json:"createTime"`
	Groups      []string  `json:"groups,omitempty"`
	Password    string    `json:"password,omitempty"`
}

func (u *User) GetName() string {
	return u.Name
}

func (u *User) GetUID() string {
	return u.UID
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("username can not be empty")
	}

	if u.Password == "" {
		return errors.New("password can not be empty")
	}

	return nil
}

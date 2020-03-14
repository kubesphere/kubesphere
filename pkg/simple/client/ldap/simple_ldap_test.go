package ldap

import (
    "github.com/google/go-cmp/cmp"
    "kubesphere.io/kubesphere/pkg/api/iam"
    "testing"
    "time"
)

func TestSimpleLdap(t *testing.T) {
    ldapClient := NewSimpleLdap()

    foo := &iam.User{
        Username:    "jerry",
        Email:       "jerry@kubesphere.io",
        Lang:        "en",
        Description: "Jerry is kind and gentle.",
        CreateTime:  time.Now(),
        Groups:      []string{},
        Password:    "P@88w0rd",
    }

    t.Run("should create user", func(t *testing.T) {
        err := ldapClient.Create(foo)
        if err != nil {
            t.Fatal(err)
        }

        // check if user really created
        user, err := ldapClient.Get(foo.Username)
        if err != nil {
            t.Fatal(err)
        }
        if diff := cmp.Diff(user, foo); len(diff) != 0 {
            t.Fatalf("%T differ (-got, +want): %s", user, diff)
        }

        _ = ldapClient.Delete(foo.Username)
    })

    t.Run("should update user", func(t *testing.T) {
        err := ldapClient.Create(foo)
        if err != nil {
            t.Fatal(err)
        }

        foo.Description = "Jerry needs some drinks."
        err = ldapClient.Update(foo)
        if err != nil {
            t.Fatal(err)
        }

        // check if user really created
        user, err := ldapClient.Get(foo.Username)
        if err != nil {
            t.Fatal(err)
        }
        if diff := cmp.Diff(user, foo); len(diff) != 0 {
            t.Fatalf("%T differ (-got, +want): %s", user, diff)
        }

        _ = ldapClient.Delete(foo.Username)
    })

    t.Run("should delete user", func(t *testing.T) {
        err := ldapClient.Create(foo)
        if err != nil {
            t.Fatal(err)
        }

        err = ldapClient.Delete(foo.Username)
        if err != nil {
            t.Fatal(err)
        }

        _, err = ldapClient.Get(foo.Username)
        if err == nil || err != ErrUserNotExists {
            t.Fatalf("expected ErrUserNotExists error, got %v", err)
        }
    })

    t.Run("should verify username and password", func(t *testing.T) {
        err := ldapClient.Create(foo)
        if err != nil {
            t.Fatal(err)
        }

        err = ldapClient.Verify(foo.Username, foo.Password)
        if err != nil {
            t.Fatalf("should pass but got an error %v", err)
        }

        err = ldapClient.Verify(foo.Username, "gibberish")
        if err == nil || err != ErrInvalidCredentials {
            t.Fatalf("expected error ErrInvalidCrenentials but got %v", err)
        }
    })
}
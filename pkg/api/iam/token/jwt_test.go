package token

import (
	"github.com/google/go-cmp/cmp"
	"kubesphere.io/kubesphere/pkg/api/iam"
	"testing"
)

func TestJwtTokenIssuer(t *testing.T) {
	issuer := NewJwtTokenIssuer(DefaultIssuerName, []byte("kubesphere"))

	testCases := []struct {
		description string
		name        string
		email       string
	}{
		{
			name:  "admin",
			email: "admin@kubesphere.io",
		},
		{
			name:  "bar",
			email: "bar@kubesphere.io",
		},
	}

	for _, testCase := range testCases {
		user := &iam.User{
			Username: testCase.name,
			Email:    testCase.email,
		}

		t.Run(testCase.description, func(t *testing.T) {
			token, err := issuer.IssueTo(user)
			if err != nil {
				t.Fatal(err)
			}

			got, err := issuer.Verify(token)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(user, got); len(diff) != 0 {
				t.Errorf("%T differ (-got, +expected), %s", user, diff)
			}
		})
	}
}

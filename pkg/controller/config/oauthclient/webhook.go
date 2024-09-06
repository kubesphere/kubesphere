/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package oauthclient

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
)

var once sync.Once

type WebhookHandler struct {
	client.Client
	getter oauth.ClientGetter
}

func (v *WebhookHandler) Default(_ context.Context, secret *v1.Secret) error {
	oc, err := oauth.UnmarshalFrom(secret)
	if err != nil {
		return err
	}
	if oc.GrantMethod == "" {
		oc.GrantMethod = oauth.GrantMethodAuto
	}
	if oc.Secret == "" {
		oc.Secret = generatePassword(32)
	}
	if secret.Labels == nil {
		secret.Labels = make(map[string]string)
	}
	secret.Labels[oauth.SecretTypeOAuthClient] = oc.Name
	return oauth.MarshalInto(oc, secret)
}

func (v *WebhookHandler) ValidateCreate(ctx context.Context, secret *corev1.Secret) (admission.Warnings, error) {
	oc, err := oauth.UnmarshalFrom(secret)
	if err != nil {
		return nil, err
	}
	if oc.Name != "" {
		exist, err := v.clientExist(ctx, oc.Name)
		if err != nil {
			return nil, err
		}
		if exist {
			return nil, fmt.Errorf("invalid OAuth client, client name '%s' already exists", oc.Name)
		}
	}

	return validate(oc)
}

func (v *WebhookHandler) ValidateUpdate(_ context.Context, old, new *corev1.Secret) (admission.Warnings, error) {
	newOc, err := oauth.UnmarshalFrom(new)
	if err != nil {
		return nil, err
	}
	oldOc, err := oauth.UnmarshalFrom(old)
	if err != nil {
		return nil, err
	}
	if newOc.Name != oldOc.Name {
		return nil, fmt.Errorf("cannot change client name")
	}
	return validate(newOc)
}

func (v *WebhookHandler) ValidateDelete(_ context.Context, _ *corev1.Secret) (admission.Warnings, error) {
	return nil, nil
}

func (v *WebhookHandler) ConfigType() corev1.SecretType {
	return oauth.SecretTypeOAuthClient
}

// validate performs general validation for the OAuth client.
func validate(oc *oauth.Client) (admission.Warnings, error) {
	if oc.Name == "" {
		return nil, fmt.Errorf("invalid OAuth client, please ensure that the client name is not empty")
	}

	if err := oauth.ValidateClient(*oc); err != nil {
		return nil, err
	}

	// Other scope values MAY be present.
	// Scope values used that are not understood by an implementation SHOULD be ignored.
	if !oauth.IsValidScopes(oc.ScopeRestrictions) {
		warnings := fmt.Sprintf("some requested scopes were invalid: %v", oc.ScopeRestrictions)
		return []string{warnings}, nil
	}
	return nil, nil
}

func (v *WebhookHandler) clientExist(ctx context.Context, clientName string) (bool, error) {
	once.Do(func() {
		v.getter = oauth.NewOAuthClientGetter(v.Client)
	})

	if _, err := v.getter.GetOAuthClient(ctx, clientName); err != nil {
		if errors.Is(err, oauth.ErrorClientNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func generatePassword(length int) string {
	characters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	password := make([]byte, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range password {
		password[i] = characters[r.Intn(len(characters))]
	}
	return string(password)
}

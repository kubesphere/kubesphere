package runners

import (
	"context"
	"errors"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fakeClientWrapper struct {
	client.Client
}

func NewFakeClientWrapper(c client.Client) client.Client {
	return &fakeClientWrapper{
		Client: c,
	}
}

func (c *fakeClientWrapper) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return errors.New("occured fake error")
}

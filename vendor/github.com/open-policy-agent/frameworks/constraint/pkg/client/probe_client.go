package client

import (
	"context"
	"fmt"
	"sort"

	"github.com/open-policy-agent/frameworks/constraint/pkg/client/drivers"
)

type Probe struct {
	client *Client
}

func NewProbe(d drivers.Driver) (*Probe, error) {
	b, err := NewBackend(Driver(d))
	if err != nil {
		return nil, err
	}
	c, err := b.NewClient(Targets(&handler{}))
	if err != nil {
		return nil, err
	}
	return &Probe{client: c}, nil
}

func (p *Probe) TestFuncs() map[string]func() error {
	ret := make(map[string]func() error)
	var names []string
	for n := range e2eTests {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		ret[name] = p.runTest(name)
	}
	return ret
}

// This must be a separate function to create a separate closure for each test
func (p *Probe) runTest(name string) func() error {
	return func() error {
		if err := p.client.Reset(context.Background()); err != nil {
			return err
		}
		err := e2eTests[name](p.client)
		if err != nil {
			dump, err2 := p.client.Dump(context.Background())
			if err2 != nil {
				dump = err2.Error()
			}
			return fmt.Errorf("error: %s\n\nOPA Dump: %s\n", err, dump) //nolint:golint
		}
		return nil
	}
}

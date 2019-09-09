package estransport

import (
	"container/ring"
	"errors"
	"net/url"
	"sync"
)

// Selector defines the interface for selecting URLs for performing request.
//
type Selector interface {
	Select() (*url.URL, error)
}

// RoundRobinSelector implements a round-robin selection strategy.
//
type RoundRobinSelector struct {
	sync.Mutex
	ring *ring.Ring
}

// Select returns a URL or error from the list of URLs in a round-robin fashion.
//
func (r *RoundRobinSelector) Select() (*url.URL, error) {
	r.Lock()
	defer r.Unlock()

	if r.ring.Len() < 1 {
		return nil, errors.New("No URL available")
	}

	v := r.ring.Value
	if ov, ok := v.(*url.URL); !ok || ov == nil {
		return nil, errors.New("No URL available")
	}

	r.ring = r.ring.Next()
	return v.(*url.URL), nil
}

// NewRoundRobinSelector creates a new RoundRobinSelector.
//
func NewRoundRobinSelector(urls ...*url.URL) *RoundRobinSelector {
	r := RoundRobinSelector{}

	r.ring = ring.New(len(urls))
	for _, u := range urls {
		r.ring.Value = u
		r.ring = r.ring.Next()
	}

	return &r
}

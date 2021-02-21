package cas

import (
	"sync"
)

// MemoryStore implements the TicketStore interface storing ticket data in memory.
type MemoryStore struct {
	mu    sync.RWMutex
	store map[string]*AuthenticationResponse
}

// Read returns the AuthenticationResponse for a ticket
func (s *MemoryStore) Read(id string) (*AuthenticationResponse, error) {
	s.mu.RLock()

	if s.store == nil {
		s.mu.RUnlock()
		return nil, ErrInvalidTicket
	}

	t, ok := s.store[id]
	s.mu.RUnlock()

	if !ok {
		return nil, ErrInvalidTicket
	}

	return t, nil
}

// Write stores the AuthenticationResponse for a ticket
func (s *MemoryStore) Write(id string, ticket *AuthenticationResponse) error {
	s.mu.Lock()

	if s.store == nil {
		s.store = make(map[string]*AuthenticationResponse)
	}

	s.store[id] = ticket

	s.mu.Unlock()
	return nil
}

// Delete removes the AuthenticationResponse for a ticket
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	delete(s.store, id)
	s.mu.Unlock()
	return nil
}

// Clear removes all ticket data
func (s *MemoryStore) Clear() error {
	s.mu.Lock()
	s.store = nil
	s.mu.Unlock()
	return nil
}

package cas

import "sync"

// SessionStore store the session's ticket
// SessionID is retrived from cookies
type SessionStore interface {
	// Get the ticket with the session id
	Get(sessionID string) (string, bool)

	// Set the session with a ticket
	Set(sessionID, ticket string) error

	// Delete the session
	Delete(sessionID string) error
}

// NewMemorySessionStore create a default SessionStore that uses memory
func NewMemorySessionStore() SessionStore {
	return &memorySessionStore{
		sessions: make(map[string]string),
	}
}

type memorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]string
}

func (m *memorySessionStore) Get(sessionID string) (string, bool) {
	m.mu.RLock()
	ticket, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	return ticket, ok
}

func (m *memorySessionStore) Set(sessionID, ticket string) error {
	m.mu.Lock()
	m.sessions[sessionID] = ticket
	m.mu.Unlock()

	return nil
}

func (m *memorySessionStore) Delete(sessionID string) error {
	m.mu.Lock()
	delete(m.sessions, sessionID)
	m.mu.Unlock()

	return nil
}

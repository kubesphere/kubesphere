package cas

import (
	"errors"
)

// TicketStore errors
var (
	// Given Ticket is not associated with an AuthenticationResponse
	ErrInvalidTicket = errors.New("cas: ticket store: invalid ticket")
)

// TicketStore provides an interface for storing and retrieving service
// ticket data.
type TicketStore interface {
	// Read returns the AuthenticationResponse data associated with a ticket identifier.
	Read(id string) (*AuthenticationResponse, error)

	// Write stores the AuthenticationResponse data received from a ticket validation.
	Write(id string, ticket *AuthenticationResponse) error

	// Delete removes the AuthenticationResponse data associated with a ticket identifier.
	Delete(id string) error

	// Clear removes all of the AuthenticationResponse data from the store.
	Clear() error
}

package token

type User interface {
	// Name
	Name() string

	UID() string
}

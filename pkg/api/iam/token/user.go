package token

type User interface {
	// Name
	GetName() string

	// UID
	GetUID() string
}

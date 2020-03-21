package token

type User interface {
	// Name
	GetName() string

	// UID
	GetUID() string

	// Groups
	GetGroups() []string
}

type AuthUser struct {
	Name   string
	UID    string
	Groups []string
}

func (a AuthUser) GetName() string {
	return a.Name
}

func (a AuthUser) GetUID() string {
	return a.UID
}

func (a AuthUser) GetGroups() []string {
	return a.Groups
}

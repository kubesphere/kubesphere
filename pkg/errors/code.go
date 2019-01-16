package errors

type Code int

const (
	OK = iota
	Canceled
	Unknown
	InvalidArgument
	Internal
	NotFound
)

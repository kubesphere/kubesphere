package errors

//go:generate stringer -type=Code
type Code int

const (
	OK Code = iota
	Canceled
	Unknown
	InvalidArgument
	Internal // 5
	Unavailable
	AlreadyExists
	WTF
	NotFound
	NotImplement
	VerifyFailed
	Conflict
)

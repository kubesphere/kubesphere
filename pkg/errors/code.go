package errors

type Code int

const (
	OK Code = iota
	Canceled
	Unknown
	InvalidArgument
	Internal // 5
	Unavailable
	AlreadyExists
	NotFound
	NotImplement
	VerifyFailed
	Conflict
)

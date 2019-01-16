package errors

//go:generate stringer -type=Code
type Code int

const (
	OK              Code = 0
	Unknown         Code = -1
	InvalidArgument Code = 4000
	Internal        Code = 5000
	Unauthorized         = 4010
	Forbidden            = 4030
	Unavailable     Code = 5030
	WTF             Code = 4180
	NotFound        Code = 4040
	NotImplement    Code = 5010
	Conflict             = 4090
)

package errors

import "strconv"

type Error struct {
	Code int
	Msg  string
}

func (e Error) Error() string {
	return strconv.Itoa(e.Code) + ": " + e.Msg
}

func New(code int, msg string) Error {
	return Error{
		Code: code,
		Msg:  msg,
	}
}

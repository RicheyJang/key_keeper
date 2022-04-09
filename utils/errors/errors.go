package errors

import (
	"fmt"
	"strconv"
)

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

func Newf(code int, format string, args ...interface{}) Error {
	return New(code, fmt.Sprintf(format, args...))
}

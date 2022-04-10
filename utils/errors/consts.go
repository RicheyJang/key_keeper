package errors

const (
	CodeInner   = 10000
	CodeKey     = 10001
	CodeRequest = 10002

	CodeWrongPasswd = 10004
	CodeUserFrozen  = 10005
	CodeNeedLogin   = 10006
)

var (
	Unknown        = New(CodeInner, "unknown error")
	NoSuchKey      = New(CodeKey, "no such key")
	InvalidRequest = New(CodeRequest, "invalid request")
	InvalidKeeper  = New(CodeRequest, "invalid keeper")
	WrongPasswd    = New(CodeWrongPasswd, "wrong password")
	UserFrozen     = New(CodeUserFrozen, "user is frozen")
	InvalidToken   = New(CodeNeedLogin, "invalid token")
)

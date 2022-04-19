package errors

const (
	CodeInner   = 10000
	CodeKey     = 10001
	CodeRequest = 10002

	CodePermission     = 10003
	CodeWrongPasswd    = 10004
	CodeUserFrozen     = 10005
	CodeNeedLogin      = 10006
	CodeUserExist      = 10007
	CodeInstanceExist  = 10008
	CodeInstanceFrozen = 10009
	CodeKeeperSupport  = 10010
)

var (
	Unknown          = New(CodeInner, "unknown error")
	NoSuchKey        = New(CodeKey, "no such key")
	InvalidRequest   = New(CodeRequest, "invalid request")
	InvalidKeeper    = New(CodeRequest, "invalid keeper")
	WrongPasswd      = New(CodeWrongPasswd, "wrong password")
	UserFrozen       = New(CodeUserFrozen, "user is frozen")
	InvalidToken     = New(CodeNeedLogin, "invalid token")
	PermissionDeny   = New(CodePermission, "permission deny")
	UserExist        = New(CodeUserExist, "user already exist")
	NoSuchInstance   = New(CodeRequest, "no such instance")
	InstanceExist    = New(CodeInstanceExist, "instance identifier already exist")
	InstanceFrozen   = New(CodeInstanceFrozen, "current instance has been frozen")
	KeeperNotSupport = New(CodeKeeperSupport, "current keeper not support this operation")
)

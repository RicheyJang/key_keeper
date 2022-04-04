package errors

const CodeInner = 10000

var (
	Unknown   = New(CodeInner, "unknown error")
	NoSuchKey = New(10001, "no such key")
)

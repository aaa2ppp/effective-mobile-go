package model

import "fmt"

type Error struct {
	code int
	msg  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d %s", e.code, e.msg)
}

func (e *Error) Code() int {
	return e.code
}

var (
	ErrBadRequest    = &Error{400, "bad request"}
	ErrNotFound      = &Error{404, "not fond"}
	ErrInternalError = &Error{500, "internal error"}
)

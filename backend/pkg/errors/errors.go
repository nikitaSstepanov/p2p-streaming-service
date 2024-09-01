package errors

import "net/http"

type Error struct {
	Message string     `json:"error"`
	Code    StatusType `json:"-"`
}

type StatusType int

const (
	Internal StatusType = iota
	NotFound
	BadInput
	Conflict
	Forbidden
	Unauthorize
)

func New(msg string, status StatusType) *Error {
	return &Error{
		Message: msg,
		Code:   status,
	}
}

func (e *Error) ToHttpCode() int {
	switch e.Code {

	case Internal:
		return http.StatusInternalServerError

	case NotFound:
		return http.StatusNotFound

	case BadInput:
		return http.StatusBadRequest

	case Conflict:
		return http.StatusConflict

	case Unauthorize:
		return http.StatusUnauthorized

	case Forbidden:
		return http.StatusForbidden

	default:
		return http.StatusOK

	}
}
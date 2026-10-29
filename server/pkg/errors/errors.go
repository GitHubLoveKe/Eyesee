package errors

import "fmt"

type AppError struct {
	HTTPStatus int
	Code       int
	Message    string
	Cause      error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func New(code int, httpStatus int, message string) *AppError {
	return &AppError{Code: code, HTTPStatus: httpStatus, Message: message}
}

func Wrap(err error, code int, httpStatus int, message string) *AppError {
	return &AppError{Code: code, HTTPStatus: httpStatus, Message: message, Cause: err}
}

var (
	ErrBadRequest       = New(40001, 400, "bad request")
	ErrNotFound         = New(40002, 404, "resource not found")
	ErrConflict         = New(40901, 409, "resource conflict")
	ErrUnauthorized     = New(40101, 401, "unauthorized")
	ErrForbidden        = New(40301, 403, "forbidden")
	ErrInternal         = New(50001, 500, "internal server error")
	ErrDependentExists  = New(42202, 422, "dependent resources exist")
	ErrValidationFailed = New(42201, 422, "validation failed")
)

// Package errors provides a way to return msged information
// for an request error. The error is normally JSON encoded.
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/pkg/errors"
)

// Define alias
var (
	WithStack = errors.WithStack
	Wrap      = errors.Wrap
	Wrapf     = errors.Wrapf
	Is        = errors.Is
	Errorf    = errors.Errorf
)

// Error Customize the error structure for implementation errors.Error interface
type Error struct {
	Code       int32  `json:"code,omitempty"`
	Msg        string `json:"msg,omitempty"`
	Status     string `json:"status,omitempty"`
	OriginBody string `json:"origin_body,omitempty"`
}

func (e *Error) Error() string {
	return e.Msg
}

func (e *Error) WithBody(body []byte) *Error {
	e.OriginBody = string(body)
	return e
}

// New generates a custom error.
func New(id, msg string, code int32) *Error {
	return &Error{
		Code:   code,
		Msg:    msg,
		Status: http.StatusText(int(code)),
	}
}

// BadRequest generates a 400 error.
func BadRequest(format string, a ...interface{}) *Error {
	return &Error{
		Code:   http.StatusBadRequest,
		Msg:    "rpc:" + fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusBadRequest),
	}
}

// Unauthorized generates a 401 error.
func Unauthorized(format string, a ...interface{}) *Error {
	return &Error{
		Code:   http.StatusUnauthorized,
		Msg:    "rpc:" + fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusUnauthorized),
	}
}

// Forbidden generates a 403 error.
func Forbidden(format string, a ...interface{}) *Error {
	return &Error{
		Code:   http.StatusForbidden,
		Msg:    "rpc:" + fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusForbidden),
	}
}

// NotFound generates a 404 error.
func NotFound(format string, a ...interface{}) *Error {
	return &Error{
		Code:   http.StatusNotFound,
		Msg:    "rpc:" + fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusNotFound),
	}
}

// MethodNotAllowed generates a 405 error.
func MethodNotAllowed(format string, a ...interface{}) *Error {
	return &Error{
		Code:   http.StatusMethodNotAllowed,
		Msg:    "rpc:" + fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusMethodNotAllowed),
	}
}

// TooManyRequests generates a 429 error.
func TooManyRequests(format string, a ...interface{}) *Error {
	return &Error{
		Code:   http.StatusTooManyRequests,
		Msg:    "rpc:" + fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusTooManyRequests),
	}
}

// Timeout generates a 408 error.
func Timeout(format string, a ...interface{}) *Error {
	return &Error{
		Code:   http.StatusRequestTimeout,
		Msg:    "rpc:" + fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusRequestTimeout),
	}
}

// Conflict generates a 409 error.
func Conflict(format string, a ...interface{}) *Error {
	return &Error{
		Code:   http.StatusConflict,
		Msg:    "rpc:" + fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusConflict),
	}
}

// RequestEntityTooLarge generates a 413 error.
func RequestEntityTooLarge(format string, a ...interface{}) *Error {
	return &Error{
		Code:   http.StatusRequestEntityTooLarge,
		Msg:    "rpc:" + fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusRequestEntityTooLarge),
	}
}

// InternalServerError generates a 500 error.
func InternalServerError(format string, a ...interface{}) *Error {
	return &Error{
		Code:   http.StatusInternalServerError,
		Msg:    "rpc:" + fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

// Equal tries to compare errors
func Equal(err1 error, err2 error) bool {
	verr1, ok1 := err1.(*Error)
	verr2, ok2 := err2.(*Error)

	if ok1 != ok2 {
		return false
	}

	if !ok1 {
		return err1 == err2
	}

	if verr1.Code != verr2.Code {
		return false
	}

	return true
}

// As finds the first error in err's chain that matches *Error
func As(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var merr *Error
	if errors.As(err, &merr) {
		return merr, true
	}
	return nil, false
}

type MultiError struct {
	lock   *sync.Mutex
	Errors []error
}

func NewMultiError() *MultiError {
	return &MultiError{
		lock:   &sync.Mutex{},
		Errors: make([]error, 0),
	}
}

func (e *MultiError) Append(err error) {
	e.Errors = append(e.Errors, err)
}

func (e *MultiError) AppendWithLock(err error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.Append(err)
}

func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

func (e *MultiError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

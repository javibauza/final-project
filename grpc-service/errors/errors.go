package errors

import (
	"errors"
	"fmt"
	"strings"
)

const ErrUserNotFound = "user not found"
const ErrWrongPassword = "wrong password"
const ErrNoFieldsForUpdate = "no fields for update"
const ErrInvalidRequestType = "invalid request type"

type ErrNotFound struct {
	Err error
}

type ErrInvalidArgument struct {
	Err error
}

type ErrPermissionDenied struct {
	Err error
}

func (r *ErrNotFound) Error() string {
	return fmt.Sprintf("%v", r.Err)
}
func NewErrNotFound() *ErrNotFound {
	return &ErrNotFound{Err: errors.New(ErrUserNotFound)}
}

func (r *ErrInvalidArgument) Error() string {
	return fmt.Sprintf("%v", r.Err)
}
func NewErrInvalidArgument(message string) *ErrInvalidArgument {
	return &ErrInvalidArgument{Err: errors.New(message)}
}

func (r *ErrPermissionDenied) Error() string {
	return fmt.Sprintf("%v", r.Err)
}
func NewErrPermissionDenied(message string) *ErrPermissionDenied {
	return &ErrPermissionDenied{Err: errors.New(message)}
}

var ErrRequiredFields = func(fields ...string) string {
	if len(fields) > 1 {
		return strings.Join(fields, ", ") + " are required"
	} else {
		return fields[0] + " is required"
	}
}

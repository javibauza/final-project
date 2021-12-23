package errors

import (
	"errors"
	"fmt"
	"strings"
)

const ErrNoFieldsForUpdate = "no fields for update"
const ErrUnexpected = "unexpected error"
const ErrInvalidInputType = "invalid input type"

type ErrInternal struct {
	Err error
}
type ErrBadRequest struct {
	Err error
}
type ErrNotFound struct {
	Err error
}
type ErrForbidden struct {
	Err error
}

func (r ErrInternal) Error() string {
	return fmt.Sprintf("%v", r.Err)
}

func (r ErrBadRequest) Error() string {
	return fmt.Sprintf("%v", r.Err)
}
func NewErrBadRequest(message string) ErrBadRequest {
	return ErrBadRequest{Err: errors.New(message)}
}

func (r ErrNotFound) Error() string {
	return fmt.Sprintf("%v", r.Err)
}

func (r ErrForbidden) Error() string {
	return fmt.Sprintf("%v", r.Err)
}

var ErrRequiredFields = func(fields ...string) string {
	if len(fields) > 1 {
		return strings.Join(fields, ", ") + " are required"
	} else {
		return fields[0] + " is required"
	}
}

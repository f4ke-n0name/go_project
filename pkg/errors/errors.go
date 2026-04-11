package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrPermissionDenied = errors.New("permission denied")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidInput     = errors.New("invalid input")
)

func Wrap(err error, message string) error {
	return fmt.Errorf(message, err)
}

var Is = errors.Is
var As = errors.As

package models

import "errors"

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
	ErrMismatch      = errors.New("mismatch")
	ErrNotEnough     = errors.New("not enough")
)

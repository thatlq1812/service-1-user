package repository

import "errors"

var (
	// ErrUserNotFound
	ErrUserNotFound = errors.New("user not found")

	// ErrEmailDuplicate
	ErrEmailDuplicate = errors.New("email already exists")
)

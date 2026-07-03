package auth

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrLastAdmin        = errors.New("cannot modify last admin")
	ErrUserDisabled     = errors.New("user disabled")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrEmptyPatch       = errors.New("empty patch")
)

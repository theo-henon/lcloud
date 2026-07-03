package dashboard

import "errors"

var (
	ErrInvalidLayout    = errors.New("invalid layout")
	ErrUnknownWidget    = errors.New("unknown widget type")
	ErrWidgetNotAllowed = errors.New("widget not allowed for role")
)

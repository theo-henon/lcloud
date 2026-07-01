package plugin

import "errors"

var (
	ErrPluginNotFound     = errors.New("plugin not found")
	ErrPluginStartFailed  = errors.New("plugin start failed")
	ErrPluginNotRunning   = errors.New("plugin not running")
)

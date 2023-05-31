package filterql

import "errors"

var (
	ErrUnexpectedEnd   = errors.New("unexpected end")
	ErrUnexpectedToken = errors.New("unexpected token")
	ErrTypeNotMatched  = errors.New("type not match")
	ErrNoSuchMethod    = errors.New("no such method")
)

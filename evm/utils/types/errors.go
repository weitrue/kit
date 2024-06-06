package types

import "errors"

var (
	ErrSize               = errors.New("data size err")
	ErrUnsupportedStorage = errors.New("unsupported storage")
)

package storage

import "errors"

var (
	ErrURLNotFound = errors.New("url not found")
	ErrAliasExists = errors.New("alias exists")
)

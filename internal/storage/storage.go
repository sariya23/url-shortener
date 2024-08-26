package storage

import "errors"

var (
	ErrURLNotFound = errors.New("url now found")
	ErrURLExists   = errors.New("url exists")
)

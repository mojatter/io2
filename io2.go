// Package io2 provides utilities for the "io" and "io/fs" package.
package io2

import (
	"errors"
	"io/fs"
	"os"
)

var (
	// ErrNotImplemented "not implemented"
	ErrNotImplemented = errors.New("not implemented")
)

var osOpen = os.Open

var fsStat = func(file *os.File) (fs.FileInfo, error) {
	return file.Stat()
}

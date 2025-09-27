package webdav

import "errors"

var (
	ErrFailedMkdirAll = errors.New("failed to create directory")
	ErrFailedDownload = errors.New("failed to download file")
	ErrFailedWrite    = errors.New("failed to write file")
	ErrReadFile       = errors.New("failed to read file")
)

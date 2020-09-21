package domain

import (
	"errors"
	"fmt"

	"git.ronaksoft.com/river/msg/msg"
)

var (
	ErrRequestTimeout        = errors.New("request timeout")
	ErrInvalidConstructor    = errors.New("constructor did not expected")
	ErrSecretNonceMismatch   = errors.New("secret hash does not match")
	ErrAuthFailed            = errors.New("creating auth key failed")
	ErrNoConnection          = errors.New("no connection")
	ErrNotFound              = errors.New("not found")
	ErrDoesNotExists         = errors.New("does not exists")
	ErrQueuePathIsNotSet     = errors.New("queue path is not set")
	ErrInvalidUserMessageKey = errors.New("invalid user message key")
	ErrAlreadyDownloading    = errors.New("already is downloading")
	ErrAlreadyUploading      = errors.New("already is uploading")
	ErrAlreadyUploaded       = errors.New("already uploaded")
	ErrNoFilePath            = errors.New("no file path")
	ErrInvalidData           = errors.New("invalid data")
	ErrServer                = errors.New("server error")
	ErrFileTooLarge          = errors.New("file is too large")
	ErrNoPostProcess         = errors.New("no post process")
)

// ParseServerError ...
func ParseServerError(b []byte) error {
	x := new(msg.Error)
	_ = x.Unmarshal(b)
	return fmt.Errorf("%s:%s", x.Code, x.Items)
}

package domain

import (
	"errors"
	"fmt"

	msg "git.ronaksoftware.com/river/msg/chat"
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
	ErrMaxFileSize           = errors.New("max file size limit")
	ErrUnknownFileSize       = errors.New("unknown file size")
	ErrAlreadyDownloading    = errors.New("already is downloading")
	ErrNoFilePath            = errors.New("no file path")
	ErrInvalidData           = errors.New("invalid data")
)

// ParseServerError ...
func ParseServerError(b []byte) error {
	x := new(msg.Error)
	_ = x.Unmarshal(b)
	return fmt.Errorf("%s:%s", x.Code, x.Items)
}

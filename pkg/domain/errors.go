package domain

import (
	"errors"
	"fmt"

	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
)

var (
	ErrHandlerNotSet         = errors.New("handlers are not set")
	ErrRequestTimeout        = errors.New("request timeout")
	ErrInvalidConstructor    = errors.New("constructor did not expected")
	ErrSecretNonceMismatch   = errors.New("secret hash does not match")
	ErrAuthFailed            = errors.New("creating auth key failed")
	ErrNoConnection          = errors.New("no connection")
	ErrNotFound              = errors.New("not found")
	ErrDoesNotExists         = errors.New("does not exists")
	ErrQueuePathIsNotSet     = errors.New("queue path is not set")
	ErrInvalidUserMessageKey = errors.New("invalid user message key")
	ErrNilDialog             = errors.New("nil dialog")
)

// ParseServerError ...
func ParseServerError(b []byte) error {
	x := new(msg.Error)
	_ = x.Unmarshal(b)
	return fmt.Errorf("%s:%s", x.Code, x.Items)
}

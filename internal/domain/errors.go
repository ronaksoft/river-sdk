package domain

import (
	"errors"
	"github.com/ronaksoft/rony"
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
	ErrLimitReached          = errors.New("limit reached")
	ErrServer                = errors.New("server error")
	ErrFileTooLarge          = errors.New("file is too large")
	ErrNoPostProcess         = errors.New("no post process")
)

// ParseServerError ...
func ParseServerError(b []byte) error {
	x := &rony.Error{}
	_ = x.Unmarshal(b)
	return x
}

func CheckErrorCode(err *rony.Error, code string) bool {
	if err.Code == code {
		return true
	}
	return false
}
func CheckError(err *rony.Error, code, item string) bool {
	if err.Code == code && err.Items == item {
		return true
	}
	return false
}

package entity

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrRequestFailed = errors.New("request failed")
	ErrNameInvalid   = errors.New("name invalid")
	ErrStatusInvalid = errors.New("status invalid")
	ErrUUIDInvalid   = errors.New("uuid invalid")
)

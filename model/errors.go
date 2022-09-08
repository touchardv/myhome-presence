package model

import "errors"

var (
	ErrMissingDeviceStatus = errors.New("missing device status")
	ErrInvalidDeviceStatus = errors.New("invalid device status")
)

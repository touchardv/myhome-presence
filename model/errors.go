package model

import "errors"

var (
	ErrInvalidDeviceAction = errors.New("invalid device action")
	ErrMissingDeviceStatus = errors.New("missing device status")
	ErrInvalidDeviceStatus = errors.New("invalid device status")
)

package internalErrors

import "errors"

var (
	ErrProductNotFound       = errors.New("product not found")
	ErrInvalidCity           = errors.New("invalid city")
	ErrActiveReceptionExists = errors.New("active reception exists")
	ErrNoActiveReception     = errors.New("no active reception")
	ErrEmailExists           = errors.New("email exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidProductType    = errors.New("invalid product type")
	ErrUserNotFound          = errors.New("user not found")
)

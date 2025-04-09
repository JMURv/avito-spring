package auth

import "errors"

var ErrInvalidToken = errors.New("invalid token")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrWhileCreatingToken = errors.New("error while creating token")
var ErrUnexpectedSignMethod = errors.New("unexpected signing method")

package repo

import "errors"

var ErrTypeIsNotValid = errors.New("type is not valid")
var ErrCityIsNotValid = errors.New("city is not valid")
var ErrNotFound = errors.New("not found")
var ErrReceptionAlreadyClosed = errors.New("reception already closed")
var ErrNoItems = errors.New("no items")
var ErrReceptionStillOpen = errors.New("reception still open")
var ErrNoActiveReception = errors.New("no active reception")

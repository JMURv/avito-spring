package ctrl

import "errors"

var ErrCityIsNotValid = errors.New("city is not valid")
var ErrTypeIsNotValid = errors.New("type is not valid")
var ErrReceptionAlreadyClosed = errors.New("reception already closed")
var ErrNoItems = errors.New("no items")
var ErrReceptionStillOpen = errors.New("reception still open")
var ErrNoActiveReception = errors.New("no active reception")

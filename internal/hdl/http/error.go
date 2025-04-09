package http

import "errors"

var ErrInvalidPathSegments = errors.New("missing or invalid path segments")
var ErrFailedToParseUUID = errors.New("failed to parse uuid")

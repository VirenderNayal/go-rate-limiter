package apperrors

import "errors"

var ErrRateLimited = errors.New("rate limited")
var ErrInvalidAlgorithm = errors.New("invalid algorithm")

package apperror

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrUnauthorized = errors.New("unauthorized user")
var ErrGUIDRequired = errors.New("guid required")
var ErrTypecastJWT = errors.New("failed to typecast jwt claims")
var ErrIncorrectRefreshToken = errors.New("invalid refresh token format")

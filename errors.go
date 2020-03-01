// 定义错误编号，处理错误信息
package doris

import (
	"errors"
	"net/http"
)

// Errors
var HTTPErrorMessages = map[int]error{
	http.StatusOK:                    errors.New("Success"),
	http.StatusUnsupportedMediaType:  errors.New("Unsupported mediatype"),
	http.StatusNotFound:              errors.New("Not found"),
	http.StatusUnauthorized:          errors.New("Unauthorized"),
	http.StatusForbidden:             errors.New("Forbidden"),
	http.StatusMethodNotAllowed:      errors.New("Method not allowed"),
	http.StatusRequestEntityTooLarge: errors.New("Request entity too large"),
	http.StatusTooManyRequests:       errors.New("Too many requests"),
	http.StatusBadRequest:            errors.New("Bad request"),
	http.StatusBadGateway:            errors.New("Bad gateway"),
	http.StatusInternalServerError:   errors.New("Internal server error"),
	http.StatusRequestTimeout:        errors.New("Request timeout"),
	http.StatusServiceUnavailable:    errors.New("Service unavailable"),
}

// Define jwt Errors
var (
	TokenExpiredErr     error = errors.New("Token is expired")
	TokenNotValidYetErr error = errors.New("Token not active yet")
	TokenMalformedErr   error = errors.New("That's not even a token")
	TokenInvalidErr     error = errors.New("Couldn't handle this token:")
	JWTMissingErr       error = errors.New("Missing or Malformed JWT")
	TokenRefreshErr     error = errors.New("This token is for refresh!")
)

// define jwt err code
// 10xxx is system error of the doris
var (
	TokenExpired     int = 10400
	TokenNotValidYet int = 10401
	TokenMalformed   int = 10402
	TokenInvalid     int = 10403
	JWTMissing       int = 10404
	TokenRefresh     int = 10405
)

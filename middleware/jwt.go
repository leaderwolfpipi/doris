package middleware

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/leaderwolfpipi/doris"
)

type (
	// JWTConfig defines the config for JWT middleware.
	JWTConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// SuccessHandler defines a function which is executed for a valid token.
		SuccessHandler JWTSuccessHandler

		// ErrorHandler defines a function which is executed for an invalid token.
		// It may be used to define a custom JWT error.
		ErrorHandler JWTErrorHandler

		// ErrorHandlerWithContext is almost identical to ErrorHandler, but it's passed the current context.
		ErrorHandlerWithContext JWTErrorHandlerWithContext

		// Signing key to validate token. Used as fallback if SigningKeys has length 0.
		// Required. This or SigningKeys.
		SigningKey interface{}

		// Signing method, used to check token signing method.
		// Optional. Default value HS256.
		SigningMethod string

		// Context key to store user information from the token into context.
		// Optional. Default value "user".
		ContextKey string

		// Claims are extendable claims data defining token content.
		// Optional. Default value jwt.MapClaims
		// for expand
		Claims jwt.Claims

		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "param:<name>"
		// - "cookie:<name>"
		TokenLookup string

		// AuthScheme to be used in the Authorization header.
		// Optional. Default value "Bearer".
		AuthScheme string

		// Get SigningKey func
		keyFunc jwt.Keyfunc
	}

	// Skipper defines a function to skip middleware. Returning true skips processing
	// the middleware.
	Skipper func(*doris.Context) bool

	// JWTSuccessHandler defines a function which is executed for a valid token.
	JWTSuccessHandler func(*doris.Context)

	// JWTErrorHandler defines a function which is executed for an invalid token.
	JWTErrorHandler func(error) error

	// JWTErrorHandlerWithContext is almost identical to JWTErrorHandler, but it's passed the current context.
	JWTErrorHandlerWithContext func(error, *doris.Context) error

	jwtExtractor func(*doris.Context) (string, error)
)

// Default Algorithms
const (
	AlgorithmHS256 = "HS256"
	// DefaultSigningKey = "secret"
)

var (
	// DefaultJWTConfig is the default JWT auth middleware config.
	DefaultJWTConfig = JWTConfig{
		Skipper:       DefaultSkipper,
		SigningMethod: AlgorithmHS256,
		ContextKey:    "user",
		TokenLookup:   "header:" + doris.Authorization,
		AuthScheme:    "Bearer",
		Claims:        jwt.MapClaims{},
	}
)

// JWT returns a JSON Web Token (JWT) auth middleware.
//
// For valid token, it sets the user in context and calls next handler.
// For invalid token, it returns "401 - Unauthorized" error.
// For missing token, it returns "400 - Bad Request" error.
//
// See: https://jwt.io/introduction
// See `JWTConfig.TokenLookup`
func JWT(key interface{}) doris.HandlerFunc {
	c := DefaultJWTConfig
	c.SigningKey = key
	return JWTWithConfig(c)
}

// JWTWithConfig returns a JWT auth middleware with config.
// See: `JWT()`.
func JWTWithConfig(config JWTConfig) doris.HandlerFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultJWTConfig.Skipper
	}
	if config.SigningKey == nil {
		panic("doris: jwt middleware requires signing key")
	}
	if config.SigningMethod == "" {
		config.SigningMethod = DefaultJWTConfig.SigningMethod
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultJWTConfig.ContextKey
	}
	if config.Claims == nil {
		config.Claims = DefaultJWTConfig.Claims
	}
	if config.TokenLookup == "" {
		config.TokenLookup = DefaultJWTConfig.TokenLookup
	}
	if config.AuthScheme == "" {
		config.AuthScheme = DefaultJWTConfig.AuthScheme
	}
	config.keyFunc = func(t *jwt.Token) (interface{}, error) {
		// Check the signing method
		if t.Method.Alg() != config.SigningMethod {
			return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
		}

		return config.SigningKey, nil
	}

	// Initialize
	parts := strings.Split(config.TokenLookup, ":")
	extractor := jwtFromHeader(parts[1], config.AuthScheme)
	switch parts[0] {
	case "query":
		extractor = jwtFromQuery(parts[1])
	case "param":
		extractor = jwtFromParam(parts[1])
	case "cookie":
		extractor = jwtFromCookie(parts[1])
	}

	// Return the middleware
	return func(c *doris.Context) error {
		// init param
		var code int = http.StatusUnauthorized
		var errMsg error = nil

		if config.Skipper(c) {
			c.Next()
		}

		auth, err := extractor(c)

		if err != nil {
			if config.ErrorHandler != nil {
				return config.ErrorHandler(err)
			}

			if config.ErrorHandlerWithContext != nil {
				return config.ErrorHandlerWithContext(err, c)
			}

			// Render error json and abort
			c.Json(http.StatusUnauthorized, doris.D{"code": http.StatusUnauthorized, "message": "JWT ERR: " + err.Error()})
			c.Abort()
			return err
		}
		token := new(jwt.Token)
		// Issue #647, #656
		if _, ok := config.Claims.(jwt.MapClaims); ok {
			token, err = jwt.Parse(auth, config.keyFunc)
		} else {
			t := reflect.ValueOf(config.Claims).Type().Elem()
			claims := reflect.New(t).Interface().(jwt.Claims)
			token, err = jwt.ParseWithClaims(auth, claims, config.keyFunc)
		}

		// 判断claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && claims["auth_type"].(string) == "refresh" {
			// 说明来自刷新token
			code = doris.TokenRefresh
			errMsg = doris.TokenRefreshErr
			c.Json(http.StatusUnauthorized, doris.D{"code": code, "message": "Invalid or Expired JWT: " + errMsg.Error()})
			c.Abort()
			return errMsg
		}

		if err == nil && token.Valid {
			// Store user information from token into context.
			c.SetParam(config.ContextKey, token)
			if config.SuccessHandler != nil {
				config.SuccessHandler(c)
			}
			c.Next()
			return nil
		}

		// check err type of jwt
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				code = doris.TokenMalformed
				errMsg = doris.TokenMalformedErr
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				code = doris.TokenExpired
				errMsg = doris.TokenExpiredErr
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				code = doris.TokenNotValidYet
				errMsg = doris.TokenNotValidYetErr
			} else {
				code = doris.TokenInvalid
				errMsg = doris.TokenInvalidErr
			}
		}

		if config.ErrorHandler != nil {
			return config.ErrorHandler(err)
		}

		if config.ErrorHandlerWithContext != nil {
			return config.ErrorHandlerWithContext(err, c)
		}

		// Render error json
		c.Json(http.StatusUnauthorized, doris.D{"code": code, "message": "Invalid or Expired JWT: " + errMsg.Error() + " [ origin err: " + err.Error() + " ] "})
		c.Abort()
		return err
	}
}

// jwtFromHeader returns a `jwtExtractor` that extracts token from the request header.
func jwtFromHeader(header string, authScheme string) jwtExtractor {
	return func(c *doris.Context) (string, error) {
		auth := c.Request.Header.Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:], nil
		}
		return "", doris.JWTMissingErr
	}
}

// jwtFromQuery returns a `jwtExtractor` that extracts token from the query string.
func jwtFromQuery(param string) jwtExtractor {
	return func(c *doris.Context) (string, error) {
		token := c.QueryParam(param)
		if token == "" {
			return "", doris.JWTMissingErr
		}
		return token, nil
	}
}

// jwtFromParam returns a `jwtExtractor` that extracts token from the url param string.
func jwtFromParam(param string) jwtExtractor {
	return func(c *doris.Context) (string, error) {
		token := c.Param(param)
		if token == "" {
			return "", doris.JWTMissingErr
		}
		return token.(string), nil
	}
}

// jwtFromCookie returns a `jwtExtractor` that extracts token from the named cookie.
func jwtFromCookie(name string) jwtExtractor {
	return func(c *doris.Context) (string, error) {
		cookie, err := c.Cookie(name)
		if err != nil {
			return "", doris.JWTMissingErr
		}
		return cookie, nil
	}
}

// DefaultSkipper returns false which processes the middleware.
func DefaultSkipper(*doris.Context) bool {
	return false
}

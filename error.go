package gametorch

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorKind is a short, stable label for the kind of an Error.
type ErrorKind string

// Error kinds.
const (
	ErrorKindHTTP           ErrorKind = "http"
	ErrorKindAPI            ErrorKind = "api"
	ErrorKindRateLimited    ErrorKind = "rate_limited"
	ErrorKindConfig         ErrorKind = "config"
	ErrorKindInvalidBaseURL ErrorKind = "invalid_base_url"
	ErrorKindDecode         ErrorKind = "decode"
)

// Error is returned by every fallible SDK operation. It exposes the HTTP
// status and the API's error message, plus convenience predicates.
type Error struct {
	kind       ErrorKind
	statusCode int
	message    string
	requestID  string
	attempts   int
	err        error
}

// Error implements the error interface.
func (e *Error) Error() string {
	switch e.kind {
	case ErrorKindHTTP:
		return fmt.Sprintf("HTTP transport error: %v", e.err)
	case ErrorKindAPI:
		return fmt.Sprintf("GameTorch API error (%s): %s", statusText(e.statusCode), e.message)
	case ErrorKindRateLimited:
		return fmt.Sprintf("rate limited by GameTorch after %d attempt(s)", e.attempts)
	case ErrorKindConfig:
		return fmt.Sprintf("invalid client configuration: %s", e.message)
	case ErrorKindInvalidBaseURL:
		return fmt.Sprintf("invalid base URL: %v", e.err)
	case ErrorKindDecode:
		return fmt.Sprintf("failed to decode response: %v", e.err)
	default:
		return "gametorch error"
	}
}

// Unwrap returns the underlying error, if any.
func (e *Error) Unwrap() error { return e.err }

// Kind returns the error kind.
func (e *Error) Kind() ErrorKind { return e.kind }

// StatusCode returns the HTTP status code, or 0 when the error did not
// originate from an API response.
func (e *Error) StatusCode() int { return e.statusCode }

// APIMessage returns the API's error message, or "" when not an API error.
func (e *Error) APIMessage() string {
	if e.kind == ErrorKindAPI {
		return e.message
	}
	return ""
}

// RequestID returns the correlation id echoed by the API, if any.
func (e *Error) RequestID() string { return e.requestID }

// Attempts returns the number of attempts made for a rate-limit error.
func (e *Error) Attempts() int { return e.attempts }

// IsNotFound reports whether the error is a 404 response.
func (e *Error) IsNotFound() bool { return e.statusCode == http.StatusNotFound }

// IsUnauthorized reports whether the error is a 401 response.
func (e *Error) IsUnauthorized() bool { return e.statusCode == http.StatusUnauthorized }

// IsPaymentRequired reports whether the error is a 402 response, which
// GameTorch uses for insufficient credits or a spending/entitlement limit.
func (e *Error) IsPaymentRequired() bool { return e.statusCode == http.StatusPaymentRequired }

// IsForbidden reports whether the error is a 403 response.
func (e *Error) IsForbidden() bool { return e.statusCode == http.StatusForbidden }

// IsConflict reports whether the error is a 409 response, typically an
// idempotency request_id reused with a different body.
func (e *Error) IsConflict() bool { return e.statusCode == http.StatusConflict }

// IsRateLimited reports whether the error is a 429 response or a rate-limit
// error surfaced after retries were exhausted.
func (e *Error) IsRateLimited() bool {
	return e.statusCode == http.StatusTooManyRequests || e.kind == ErrorKindRateLimited
}

// asError converts err to an *Error, wrapping unknown errors as transport
// errors.
func asError(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return &Error{kind: ErrorKindHTTP, err: err}
}

func newAPIError(status int, message, requestID string) *Error {
	return &Error{
		kind:       ErrorKindAPI,
		statusCode: status,
		message:    message,
		requestID:  requestID,
	}
}

func newHTTPError(err error) *Error {
	return &Error{kind: ErrorKindHTTP, err: err}
}

func newRateLimitedError(attempts int) *Error {
	return &Error{kind: ErrorKindRateLimited, attempts: attempts}
}

func newConfigError(message string) *Error {
	return &Error{kind: ErrorKindConfig, message: message}
}

func newInvalidBaseURLError(err error) *Error {
	return &Error{kind: ErrorKindInvalidBaseURL, err: err}
}

func newDecodeError(err error) *Error {
	return &Error{kind: ErrorKindDecode, err: err}
}

func statusText(code int) string {
	if text := http.StatusText(code); text != "" {
		return text
	}
	return fmt.Sprintf("status %d", code)
}

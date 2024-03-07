package fail

import (
	"errors"
	"fmt"
	"net/http"
)

// Error is simply a shorthand for the fail package's StatusError that includes an HTTP status code w/ your error message.
type Error StatusError

// StatusError is an error that maintains not just an error message but an HTTP
// compatible status code indicating the type/class of error. It's useful for helping
// you figure out downstream if an occurred because the user didn't have rights or
// if some record did not exist.
type StatusError struct {
	// Status is the HTTP status code that most closely describes this error.
	Status int `json:"Status"`
	// Message is the human-readable error message.
	Message string `json:"Message"`
}

// StatusCode returns the most relevant HTTP-style status code describing this type of error.
func (r StatusError) StatusCode() int {
	return r.Status
}

// Error returns the underlying error message.
func (r StatusError) Error() string {
	return r.Message
}

// New creates an error that maps directly to an HTTP status so if your method results in
// this error, it will result in the same 'status' in your HTTP response. While you can do this
// for more obscure HTTP failure statuses like "payment required", it's typically a better idea
// to use the error functions BadRequest(), PermissionDenied(), etc. as it provides proper status
// codes and results in more readable code.
func New(status int, messageFormat string, args ...any) StatusError {
	return StatusError{
		Status:  status,
		Message: fmt.Sprintf(messageFormat, args...),
	}
}

// Status looks for either a Status(), StatusCode(), or Code() method on the error to
// figure out the most appropriate HTTP status code for it. If the error doesn't have any of
// those methods then we'll just assume that it is a 500 error.
func Status(err error) int {
	var errStatus errorWithStatus
	if errors.As(err, &errStatus) {
		return errStatus.Status()
	}

	var errStatusCode errorWithStatusCode
	if errors.As(err, &errStatusCode) {
		return errStatusCode.StatusCode()
	}

	var errCode errorWithCode
	if errors.As(err, &errCode) {
		return errCode.Code()
	}

	// This is how AWS reports HTTP style errors, so make this convenient.
	var errHTTPStatusCode errorWithHTTPStatusCode
	if errors.As(err, &errHTTPStatusCode) {
		return errHTTPStatusCode.HTTPStatusCode()
	}

	return http.StatusInternalServerError
}

// Unexpected is a generic 500-style catch-all error for failures you don't know what to do with. This is
// exactly the same as calling InternalServerError(), just more concise in your code.
func Unexpected(messageFormat string, args ...any) StatusError {
	return New(http.StatusInternalServerError, messageFormat, args...)
}

// IsUnexpected returns true if the underlying HTTP status code of 'err' is 500. This will be true for any
// error you created using the Unexpected() function.
func IsUnexpected(err error) bool {
	return Status(err) == http.StatusInternalServerError
}

// InternalServerError is a generic 500-style catch-all error for failures you don't know what to do with.
func InternalServerError(messageFormat string, args ...any) StatusError {
	return Unexpected(messageFormat, args...)
}

// IsInternalServiceError returns true if the underlying HTTP status code of 'err' is 500. This will be true for any
// error you created using the InternalServiceError() function.
func IsInternalServiceError(err error) bool {
	return Status(err) == http.StatusInternalServerError
}

// BadRequest is a 400-style error that indicates that some aspect of the request was either ill-formed
// or failed validation. This could be an ill-formed function parameter, a bad HTTP body, etc.
func BadRequest(messageFormat string, args ...any) StatusError {
	return New(http.StatusBadRequest, messageFormat, args...)
}

// IsBadRequest returns true if the underlying HTTP status code of 'err' is 400. This will be true for any
// error you created using the BadRequest() function.
func IsBadRequest(err error) bool {
	return Status(err) == http.StatusBadRequest
}

// PaymentRequired is a 402-style error that indicates that you must provide payment to access
// the given resource or perform the given task. Greedy bastard :)
func PaymentRequired(messageFormat string, args ...any) StatusError {
	return New(http.StatusPaymentRequired, messageFormat, args...)
}

// IsPaymentRequired returns true if the underlying HTTP status code of 'err' is 402. This will be tru for any
// error you created using the PaymentRequired() function.
func IsPaymentRequired(err error) bool {
	return Status(err) == http.StatusPaymentRequired
}

// BadCredentials is a 401-style error that indicates that the caller either didn't provide credentials
// when necessary or they did, but the credentials were invalid for some reason. This corresponds to the
// HTTP "unauthorized" status, but we prefer this name because this type of failure has nothing to do
// with authorization, and it's more clear what aspect of the request has failed.
func BadCredentials(messageFormat string, args ...any) StatusError {
	return New(http.StatusUnauthorized, messageFormat, args...)
}

// IsBadCredentials returns true if the underlying HTTP status code of 'err' is 401. This will be true for any
// error you created using the BadCredentials() function.
func IsBadCredentials(err error) bool {
	return Status(err) == http.StatusUnauthorized
}

// PermissionDenied is a 403-style error that indicates that the caller does not have rights/clearance
// to perform any part of the operation.
func PermissionDenied(messageFormat string, args ...any) StatusError {
	return New(http.StatusForbidden, messageFormat, args...)
}

// IsPermissionDenied returns true if the underlying HTTP status code of 'err' is 403. This will be true for any
// error you created using the PermissionDenied() function.
func IsPermissionDenied(err error) bool {
	return Status(err) == http.StatusForbidden
}

// NotFound is a 404-style error that indicates that some record/resource could not be located.
func NotFound(messageFormat string, args ...any) StatusError {
	return New(http.StatusNotFound, messageFormat, args...)
}

// IsNotFound returns true if the underlying HTTP status code of 'err' is 404. This will be true for any
// error you created using the NotFound() function.
func IsNotFound(err error) bool {
	return Status(err) == http.StatusNotFound
}

// MethodNotAllowed is a 405-style error that indicates that the wrong HTTP method was used.
func MethodNotAllowed(messageFormat string, args ...any) StatusError {
	return New(http.StatusMethodNotAllowed, messageFormat, args...)
}

// IsMethodNotAllowed returns true if the underlying HTTP status code of 'err' is 405. This will be true for any
// error you created using the MethodNotAllowed() function.
func IsMethodNotAllowed(err error) bool {
	return Status(err) == http.StatusMethodNotAllowed
}

// Timeout is a 408-style error that indicates that some operation exceeded its allotted time/deadline.
func Timeout(messageFormat string, args ...any) StatusError {
	return New(http.StatusRequestTimeout, messageFormat, args...)
}

// IsTimeout returns true if the underlying HTTP status code of 'err' is 408. This will be true for any
// error you created using the Timeout() function.
func IsTimeout(err error) bool {
	return Status(err) == http.StatusRequestTimeout
}

// AlreadyExists is a 409-style error that is used when attempting to create some record/resource, but
// there is already a duplicate instance in existence.
func AlreadyExists(messageFormat string, args ...any) StatusError {
	return New(http.StatusConflict, messageFormat, args...)
}

// IsAlreadyExists returns true if the underlying HTTP status code of 'err' is 409. This will be true for any
// error you created using the AlreadyExists() function.
func IsAlreadyExists(err error) bool {
	return Status(err) == http.StatusConflict
}

// Gone is a 410-style error that is used to indicate that something used to exist, but doesn't anymore.
func Gone(messageFormat string, args ...any) StatusError {
	return New(http.StatusGone, messageFormat, args...)
}

// IsGone returns true if the underlying HTTP status code of 'err' is 410. This will be true for any
// error you created using the Gone() function.
func IsGone(err error) bool {
	return Status(err) == http.StatusGone
}

// TooLarge is a 413-style error that is used to indicate that some resource/entity is too big.
func TooLarge(messageFormat string, args ...any) StatusError {
	return New(http.StatusRequestEntityTooLarge, messageFormat, args...)
}

// IsTooLarge returns true if the underlying HTTP status code of 'err' is 413. This will be true for any
// error you created using the TooLarge() function.
func IsTooLarge(err error) bool {
	return Status(err) == http.StatusRequestEntityTooLarge
}

// UnsupportedFormat is a 415-style error that is used to indicate that the media/content type of
// some input is not valid. For instance, the user uploads an "image/bmp" but you only support
// PNG and JPG files.
func UnsupportedFormat(messageFormat string, args ...any) StatusError {
	return New(http.StatusUnsupportedMediaType, messageFormat, args...)
}

// IsUnsupportedFormat returns true if the underlying HTTP status code of 'err' is 415. This will be true for any
// error you created using the UnsupportedFormat() function.
func IsUnsupportedFormat(err error) bool {
	return Status(err) == http.StatusUnsupportedMediaType
}

// Throttled is a 429-style error that indicates that the caller has exceeded the number of requests,
// amount of resources, etc allowed over some time period. The failure should indicated to the caller
// that the failure is due to some throttle that prevented the operation from even occurring.
func Throttled(messageFormat string, args ...any) StatusError {
	return New(http.StatusTooManyRequests, messageFormat, args...)
}

// IsThrottled returns true if the underlying HTTP status code of 'err' is 429. This will be true for any
// error you created using the Throttled() function.
func IsThrottled(err error) bool {
	return Status(err) == http.StatusTooManyRequests
}

// NotImplemented is a 501-style error that indicates that the resource/logic required to fulfill
// the request has not been added yet.
func NotImplemented(messageFormat string, args ...any) StatusError {
	return New(http.StatusNotImplemented, messageFormat, args...)
}

// IsNotImplemented returns true if the underlying HTTP status code of 'err' is 501. This will be true for any
// error you created using the Throttled() function.
func IsNotImplemented(err error) bool {
	return Status(err) == http.StatusNotImplemented
}

// BadGateway is a 502-style error that indicates that some upstream resource was not available. Your
// code is working fine, but another service you're dependent on is not.
func BadGateway(messageFormat string, args ...any) StatusError {
	return New(http.StatusBadGateway, messageFormat, args...)
}

// IsBadGateway returns true if the underlying HTTP status code of 'err' is 502. This will be true for any
// error you created using the BadGateway() function.
func IsBadGateway(err error) bool {
	return Status(err) == http.StatusBadGateway
}

// Unavailable is a 503-style error that indicates that some aspect of the server/service is unavailable.
// This could be something like DB connection failures, some third party service being down, etc.
func Unavailable(messageFormat string, args ...any) StatusError {
	return New(http.StatusServiceUnavailable, messageFormat, args...)
}

// IsUnavailable returns true if the underlying HTTP status code of 'err' is 503. This will be true for any
// error you created using the Unavailable() function.
func IsUnavailable(err error) bool {
	return Status(err) == http.StatusServiceUnavailable
}

/*
 * These interfaces are used to help extract status codes from generic error instances.
 */

type errorWithStatus interface {
	error
	Status() int
}

type errorWithStatusCode interface {
	error
	StatusCode() int
}

type errorWithCode interface {
	error
	Code() int
}

type errorWithHTTPStatusCode interface {
	error
	HTTPStatusCode() int
}

// ErrorHandler is the generic function signature for something that accepts errors
// that occur in the bowels of the framework. It gives you a chance to log or deal
// with them as you see fit. These are typically asynchronous things where you don't
// have any handle over the control flow, but you don't want to lose these events.
type ErrorHandler func(err error)

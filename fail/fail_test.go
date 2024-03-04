//go:build unit

package fail_test

import (
	"fmt"
	"testing"

	"github.com/bridgekitio/frodo/fail"
	"github.com/stretchr/testify/suite"
)

type FailSuite struct {
	suite.Suite
}

func (suite *FailSuite) TestNew() {
	suite.assertError(fail.New(100, "foo"), 100, "foo")
	suite.assertError(fail.New(100, "%s", "foo"), 100, "foo")
	suite.assertError(fail.New(100, "foo %s %v", "bar", 99), 100, "foo bar 99")
}

// Since we don't return an error argument for... creating an error... we have no
// problems with you providing non-HTTP standard fail. Maybe you're doing your own
// error status mapping; who's to say.
func (suite *FailSuite) TestNew_wonkyStatus() {
	suite.assertError(fail.New(0, ""), 0, "")
	suite.assertError(fail.New(-42, ""), -42, "")
	suite.assertError(fail.New(9999, ""), 9999, "")
}

func (suite *FailSuite) TestStatus() {
	suite.Equal(400, fail.Status(fail.BadRequest("")))
	suite.Equal(400, fail.Status(fail.New(400, "")))

	suite.Equal(503, fail.Status(fail.Unavailable("")))
	suite.Equal(503, fail.Status(fail.New(503, "")))

	suite.Equal(500, fail.Status(fail.InternalServerError("")))
	suite.Equal(500, fail.Status(fail.Unexpected("")))
	suite.Equal(500, fail.Status(fail.New(500, "")))

	// No status info at all
	suite.Equal(500, fail.Status(fmt.Errorf("hello")))

	// Non-RPCError examples that do have status values
	suite.Equal(500, fail.Status(errWithCode{code: 500}))
	suite.Equal(503, fail.Status(errWithCode{code: 503}))
	suite.Equal(404, fail.Status(errWithCode{code: 404}))
	suite.Equal(500, fail.Status(errWithStatusCode{statusCode: 500}))
	suite.Equal(503, fail.Status(errWithStatusCode{statusCode: 503}))
	suite.Equal(404, fail.Status(errWithStatusCode{statusCode: 404}))
	suite.Equal(500, fail.Status(errorWithHTTPStatusCode{httpStatusCode: 500}))
	suite.Equal(503, fail.Status(errorWithHTTPStatusCode{httpStatusCode: 503}))
	suite.Equal(404, fail.Status(errorWithHTTPStatusCode{httpStatusCode: 404}))

	// Wrapping w/ %w verb should still allow status retrieval.
	suite.Equal(401, fail.Status(fmt.Errorf("wrapped failure: %w", errWithCode{code: 401})))
	suite.Equal(401, fail.Status(fmt.Errorf("wrapped failure: %w", errWithStatusCode{statusCode: 401})))
	suite.Equal(401, fail.Status(fmt.Errorf("wrapped failure: %w", errorWithHTTPStatusCode{httpStatusCode: 401})))

	var wrapped error
	wrapped = errWithCode{code: 401}
	wrapped = fmt.Errorf("wrapped failure 1: %w", wrapped)
	wrapped = fmt.Errorf("wrapped failure 2: %w", wrapped)
	wrapped = fmt.Errorf("wrapped failure 3: %w", wrapped)
	wrapped = fmt.Errorf("wrapped failure 4: %w", wrapped)
	suite.Equal(401, fail.Status(wrapped))
}

func (suite *FailSuite) TestUnexpected() {
	expectedStatus := 500
	suite.assertError(fail.Unexpected("foo"), expectedStatus, "foo")
	suite.assertError(fail.Unexpected("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.Unexpected("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsUnexpected() {
	suite.True(fail.IsUnexpected(fail.Unexpected("foo")))
	suite.True(fail.IsUnexpected(fail.InternalServerError("foo")))
	suite.True(fail.IsUnexpected(fmt.Errorf("no status present still 500")))
	suite.False(fail.IsUnexpected(fail.NotFound("")))
	suite.False(fail.IsUnexpected(fail.BadCredentials("")))

	// Non-StatusError examples
	suite.True(fail.IsUnexpected(errWithCode{code: 500}))
	suite.False(fail.IsUnexpected(errWithCode{code: 400}))
	suite.True(fail.IsUnexpected(errWithStatusCode{statusCode: 500}))
	suite.False(fail.IsUnexpected(errWithStatusCode{statusCode: 400}))
}

func (suite *FailSuite) TestInternalServerError() {
	expectedStatus := 500
	suite.assertError(fail.InternalServerError("foo"), expectedStatus, "foo")
	suite.assertError(fail.InternalServerError("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.InternalServerError("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsInternalServiceError() {
	suite.True(fail.IsInternalServiceError(fail.Unexpected("foo")))
	suite.True(fail.IsInternalServiceError(fail.InternalServerError("foo")))
	suite.True(fail.IsInternalServiceError(fmt.Errorf("no status present still 500")))
	suite.False(fail.IsInternalServiceError(fail.NotFound("")))
	suite.False(fail.IsInternalServiceError(fail.BadCredentials("")))

	// Non-RPCError examples
	suite.True(fail.IsInternalServiceError(errWithCode{code: 500}))
	suite.False(fail.IsInternalServiceError(errWithCode{code: 400}))
	suite.True(fail.IsInternalServiceError(errWithStatusCode{statusCode: 500}))
	suite.False(fail.IsInternalServiceError(errWithStatusCode{statusCode: 400}))
}

func (suite *FailSuite) TestBadRequest() {
	expectedStatus := 400
	suite.assertError(fail.BadRequest("foo"), expectedStatus, "foo")
	suite.assertError(fail.BadRequest("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.BadRequest("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsBadRequest() {
	suite.True(fail.IsBadRequest(fail.BadRequest("")))
	suite.False(fail.IsBadRequest(fail.Unexpected("")))
	suite.False(fail.IsBadRequest(fail.NotFound("")))

	// Non-RPCError examples
	suite.True(fail.IsBadRequest(errWithCode{code: 400}))
	suite.False(fail.IsBadRequest(errWithCode{code: 403}))
	suite.True(fail.IsBadRequest(errWithStatusCode{statusCode: 400}))
	suite.False(fail.IsBadRequest(errWithStatusCode{statusCode: 403}))
}

func (suite *FailSuite) TestPaymentRequired() {
	expectedStatus := 402
	suite.assertError(fail.PaymentRequired("foo"), expectedStatus, "foo")
	suite.assertError(fail.PaymentRequired("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.PaymentRequired("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsPaymentRequired() {
	suite.True(fail.IsPaymentRequired(fail.PaymentRequired("")))
	suite.False(fail.IsPaymentRequired(fail.Unexpected("")))
	suite.False(fail.IsPaymentRequired(fail.NotFound("")))

	// Non-RPCError examples
	suite.True(fail.IsPaymentRequired(errWithCode{code: 402}))
	suite.False(fail.IsPaymentRequired(errWithCode{code: 403}))
	suite.True(fail.IsPaymentRequired(errWithStatusCode{statusCode: 402}))
	suite.False(fail.IsPaymentRequired(errWithStatusCode{statusCode: 403}))
}

func (suite *FailSuite) TestBadCredentials() {
	expectedStatus := 401
	suite.assertError(fail.BadCredentials("foo"), expectedStatus, "foo")
	suite.assertError(fail.BadCredentials("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.BadCredentials("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsBadCredentials() {
	suite.True(fail.IsBadCredentials(fail.BadCredentials("")))
	suite.False(fail.IsBadCredentials(fail.Unexpected("")))
	suite.False(fail.IsBadCredentials(fail.NotFound("")))

	// Non-RPCError examples
	suite.True(fail.IsBadCredentials(errWithCode{code: 401}))
	suite.False(fail.IsBadCredentials(errWithCode{code: 403}))
	suite.True(fail.IsBadCredentials(errWithStatusCode{statusCode: 401}))
	suite.False(fail.IsBadCredentials(errWithStatusCode{statusCode: 403}))
}

func (suite *FailSuite) TestPermissionDenied() {
	expectedStatus := 403
	suite.assertError(fail.PermissionDenied("foo"), expectedStatus, "foo")
	suite.assertError(fail.PermissionDenied("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.PermissionDenied("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsPermissionDenied() {
	suite.True(fail.IsPermissionDenied(fail.PermissionDenied("")))
	suite.False(fail.IsPermissionDenied(fail.Unexpected("")))
	suite.False(fail.IsPermissionDenied(fail.NotFound("")))

	// Non-RPCError examples
	suite.True(fail.IsPermissionDenied(errWithCode{code: 403}))
	suite.False(fail.IsPermissionDenied(errWithCode{code: 404}))
	suite.True(fail.IsPermissionDenied(errWithStatusCode{statusCode: 403}))
	suite.False(fail.IsPermissionDenied(errWithStatusCode{statusCode: 404}))
}

func (suite *FailSuite) TestNotFound() {
	expectedStatus := 404
	suite.assertError(fail.NotFound("foo"), expectedStatus, "foo")
	suite.assertError(fail.NotFound("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.NotFound("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsNotFound() {
	suite.True(fail.IsNotFound(fail.NotFound("")))
	suite.False(fail.IsNotFound(fail.Unexpected("")))
	suite.False(fail.IsNotFound(fail.BadRequest("")))

	// Non-RPCError examples
	suite.True(fail.IsNotFound(errWithCode{code: 404}))
	suite.False(fail.IsNotFound(errWithCode{code: 401}))
	suite.True(fail.IsNotFound(errWithStatusCode{statusCode: 404}))
	suite.False(fail.IsNotFound(errWithStatusCode{statusCode: 401}))
}

func (suite *FailSuite) TestMethodNotFound() {
	expectedStatus := 405
	suite.assertError(fail.MethodNotAllowed("foo"), expectedStatus, "foo")
	suite.assertError(fail.MethodNotAllowed("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.MethodNotAllowed("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsMethodNotAllowed() {
	suite.True(fail.IsMethodNotAllowed(fail.MethodNotAllowed("")))
	suite.False(fail.IsMethodNotAllowed(fail.Unexpected("")))
	suite.False(fail.IsMethodNotAllowed(fail.BadRequest("")))

	// Non-RPCError examples
	suite.True(fail.IsMethodNotAllowed(errWithCode{code: 405}))
	suite.False(fail.IsMethodNotAllowed(errWithCode{code: 401}))
	suite.True(fail.IsMethodNotAllowed(errWithStatusCode{statusCode: 405}))
	suite.False(fail.IsMethodNotAllowed(errWithStatusCode{statusCode: 401}))
}

func (suite *FailSuite) TestAlreadyExists() {
	expectedStatus := 409
	suite.assertError(fail.AlreadyExists("foo"), expectedStatus, "foo")
	suite.assertError(fail.AlreadyExists("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.AlreadyExists("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsAlreadyExists() {
	suite.True(fail.IsAlreadyExists(fail.AlreadyExists("")))
	suite.False(fail.IsAlreadyExists(fail.Unexpected("")))
	suite.False(fail.IsAlreadyExists(fail.BadRequest("")))

	// Non-RPCError examples
	suite.True(fail.IsAlreadyExists(errWithCode{code: 409}))
	suite.False(fail.IsAlreadyExists(errWithCode{code: 401}))
	suite.True(fail.IsAlreadyExists(errWithStatusCode{statusCode: 409}))
	suite.False(fail.IsAlreadyExists(errWithStatusCode{statusCode: 401}))
}

func (suite *FailSuite) TestGone() {
	expectedStatus := 410
	suite.assertError(fail.Gone("foo"), expectedStatus, "foo")
	suite.assertError(fail.Gone("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.Gone("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsGone() {
	suite.True(fail.IsGone(fail.Gone("")))
	suite.False(fail.IsGone(fail.Unexpected("")))
	suite.False(fail.IsGone(fail.BadRequest("")))

	// Non-RPCError examples
	suite.True(fail.IsGone(errWithCode{code: 410}))
	suite.False(fail.IsGone(errWithCode{code: 401}))
	suite.True(fail.IsGone(errWithStatusCode{statusCode: 410}))
	suite.False(fail.IsGone(errWithStatusCode{statusCode: 401}))
}

func (suite *FailSuite) TestTooLarge() {
	expectedStatus := 413
	suite.assertError(fail.TooLarge("foo"), expectedStatus, "foo")
	suite.assertError(fail.TooLarge("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.TooLarge("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsTooLarge() {
	suite.True(fail.IsTooLarge(fail.TooLarge("")))
	suite.False(fail.IsTooLarge(fail.Unexpected("")))
	suite.False(fail.IsTooLarge(fail.BadRequest("")))

	// Non-RPCError examples
	suite.True(fail.IsTooLarge(errWithCode{code: 413}))
	suite.False(fail.IsTooLarge(errWithCode{code: 401}))
	suite.True(fail.IsTooLarge(errWithStatusCode{statusCode: 413}))
	suite.False(fail.IsTooLarge(errWithStatusCode{statusCode: 401}))
}

func (suite *FailSuite) TestUnsupportedFormat() {
	expectedStatus := 415
	suite.assertError(fail.UnsupportedFormat("foo"), expectedStatus, "foo")
	suite.assertError(fail.UnsupportedFormat("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.UnsupportedFormat("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsUnsupportedFormat() {
	suite.True(fail.IsUnsupportedFormat(fail.UnsupportedFormat("")))
	suite.False(fail.IsUnsupportedFormat(fail.Unexpected("")))
	suite.False(fail.IsUnsupportedFormat(fail.BadRequest("")))

	// Non-RPCError examples
	suite.True(fail.IsUnsupportedFormat(errWithCode{code: 415}))
	suite.False(fail.IsUnsupportedFormat(errWithCode{code: 401}))
	suite.True(fail.IsUnsupportedFormat(errWithStatusCode{statusCode: 415}))
	suite.False(fail.IsUnsupportedFormat(errWithStatusCode{statusCode: 401}))
}

func (suite *FailSuite) TestTimeout() {
	expectedStatus := 408
	suite.assertError(fail.Timeout("foo"), expectedStatus, "foo")
	suite.assertError(fail.Timeout("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.Timeout("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsTimeout() {
	suite.True(fail.IsTimeout(fail.Timeout("")))
	suite.False(fail.IsTimeout(fail.Unexpected("")))
	suite.False(fail.IsTimeout(fail.BadRequest("")))

	// Non-RPCError examples
	suite.True(fail.IsTimeout(errWithCode{code: 408}))
	suite.False(fail.IsTimeout(errWithCode{code: 401}))
	suite.True(fail.IsTimeout(errWithStatusCode{statusCode: 408}))
	suite.False(fail.IsTimeout(errWithStatusCode{statusCode: 401}))
}

func (suite *FailSuite) TestThrottled() {
	expectedStatus := 429
	suite.assertError(fail.Throttled("foo"), expectedStatus, "foo")
	suite.assertError(fail.Throttled("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.Throttled("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsThrottled() {
	suite.True(fail.IsThrottled(fail.Throttled("")))
	suite.False(fail.IsThrottled(fail.Unexpected("")))
	suite.False(fail.IsThrottled(fail.BadRequest("")))

	// Non-RPCError examples
	suite.True(fail.IsThrottled(errWithCode{code: 429}))
	suite.False(fail.IsThrottled(errWithCode{code: 401}))
	suite.True(fail.IsThrottled(errWithStatusCode{statusCode: 429}))
	suite.False(fail.IsThrottled(errWithStatusCode{statusCode: 401}))
}

func (suite *FailSuite) TestNotImplemented() {
	expectedStatus := 501
	suite.assertError(fail.NotImplemented("foo"), expectedStatus, "foo")
	suite.assertError(fail.NotImplemented("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.NotImplemented("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsNotImplemented() {
	suite.True(fail.IsNotImplemented(fail.NotImplemented("")))
	suite.False(fail.IsNotImplemented(fail.Unexpected("")))
	suite.False(fail.IsNotImplemented(fail.BadRequest("")))

	// Non-RPCError examples
	suite.True(fail.IsNotImplemented(errWithCode{code: 501}))
	suite.False(fail.IsNotImplemented(errWithCode{code: 401}))
	suite.True(fail.IsNotImplemented(errWithStatusCode{statusCode: 501}))
	suite.False(fail.IsNotImplemented(errWithStatusCode{statusCode: 401}))
}

func (suite *FailSuite) TestBadGateway() {
	expectedStatus := 502
	suite.assertError(fail.BadGateway("foo"), expectedStatus, "foo")
	suite.assertError(fail.BadGateway("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.BadGateway("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsBadGateway() {
	suite.True(fail.IsBadGateway(fail.BadGateway("")))
	suite.False(fail.IsBadGateway(fail.Unexpected("")))
	suite.False(fail.IsBadGateway(fail.BadRequest("")))

	// Non-RPCError examples
	suite.True(fail.IsBadGateway(errWithCode{code: 502}))
	suite.False(fail.IsBadGateway(errWithCode{code: 401}))
	suite.True(fail.IsBadGateway(errWithStatusCode{statusCode: 502}))
	suite.False(fail.IsBadGateway(errWithStatusCode{statusCode: 401}))
}

func (suite *FailSuite) TestUnavailable() {
	expectedStatus := 503
	suite.assertError(fail.Unavailable("foo"), expectedStatus, "foo")
	suite.assertError(fail.Unavailable("%s", "foo"), expectedStatus, "foo")
	suite.assertError(fail.Unavailable("foo %s %v", "bar", 99), expectedStatus, "foo bar 99")
}

func (suite *FailSuite) TestIsUnavailable() {
	suite.True(fail.IsUnavailable(fail.Unavailable("")))
	suite.False(fail.IsUnavailable(fail.Unexpected("")))
	suite.False(fail.IsUnavailable(fail.BadRequest("")))

	// Non-RPCError examples
	suite.True(fail.IsUnavailable(errWithCode{code: 503}))
	suite.False(fail.IsUnavailable(errWithCode{code: 401}))
	suite.True(fail.IsUnavailable(errWithStatusCode{statusCode: 503}))
	suite.False(fail.IsUnavailable(errWithStatusCode{statusCode: 401}))
}

// assertError checks that both the status and message of the resulting 'err' are what we expect.
func (suite *FailSuite) assertError(err fail.StatusError, expectedStatus int, expectedMessage string) {
	suite.Require().Equal(expectedStatus, err.StatusCode())
	suite.Require().Equal(expectedMessage, err.Error())
}

func TestFailSuite(t *testing.T) {
	suite.Run(t, new(FailSuite))
}

type errWithCode struct {
	code int
}

func (err errWithCode) Code() int {
	return err.code
}

func (err errWithCode) Error() string {
	return ""
}

type errWithStatusCode struct {
	statusCode int
}

func (err errWithStatusCode) StatusCode() int {
	return err.statusCode
}

func (err errWithStatusCode) Error() string {
	return ""
}

type errorWithHTTPStatusCode struct {
	httpStatusCode int
}

func (err errorWithHTTPStatusCode) HTTPStatusCode() int {
	return err.httpStatusCode
}

func (err errorWithHTTPStatusCode) Error() string {
	return ""
}

package clients

import (
	"net/http"

	"github.com/bridgekitio/frodo/metadata"
)

// RoundTripperFunc matches the signature of the standard http.RoundTripper interface.
type RoundTripperFunc func(r *http.Request) (*http.Response, error)

// RoundTrip allows a single round trip function to behave as a full-fledged http.RoundTripper.
func (rt RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

// ClientMiddlewareFunc is a round-tripper-like function that accepts a request and returns a response/error
// combo, but also accepts 'next' (the rest of the computation) so that you can short circuit the
// execution as you see fit.
type ClientMiddlewareFunc func(request *http.Request, next RoundTripperFunc) (*http.Response, error)

// clientMiddlewarePipeline is an ordered chain of client middleware handlers that should fire
// one after another.
type clientMiddlewarePipeline []ClientMiddlewareFunc

func (pipeline clientMiddlewarePipeline) Then(handler RoundTripperFunc) RoundTripperFunc {
	for i := len(pipeline) - 1; i >= 0; i-- {
		mw := pipeline[i]
		next := handler
		handler = func(request *http.Request) (*http.Response, error) {
			return mw(request, next)
		}
	}
	return handler

}

// writeMetadataHeader encodes all of the context's (the context on the request) metadata values as
// JSON and writes that to the "X-RPC-Values" header so that the remote service has access to all
// of your values as well.
func writeMetadataHeader(request *http.Request, next RoundTripperFunc) (*http.Response, error) {
	encodedValues := metadata.Encode(request.Context())
	request.Header.Set(metadata.Header, string(encodedValues))
	return next(request)
}

// writeAuthorizationHeader takes the authorization information on the context (if present) and applies it
// to the "Authorization" header on the request. This ensures that the credentials used to authenticate/authorize
// the request to this service are automatically applied this upstream service call, too.
func writeAuthorizationHeader(request *http.Request, next RoundTripperFunc) (*http.Response, error) {
	if auth := metadata.Authorization(request.Context()); auth != "" {
		request.Header.Set("Authorization", auth)
	}
	return next(request)
}

package events

import (
	"bytes"
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/bridgekit-io/frodo/codec"
	"github.com/bridgekit-io/frodo/eventsource"
	"github.com/bridgekit-io/frodo/fail"
	"github.com/bridgekit-io/frodo/metadata"
	"github.com/bridgekit-io/frodo/services"
)

// errorKeySuffix represents the thing we append to the end of the key we publish to the broker to indicate that this
// should route to error listeners, not success listeners.
const errorKeySuffix = ":Error"

// message is the envelope used by the event gateway to broadcast events to other services
// that might want to perform other tasks based on this event. It contains all of the information
// required for a subscriber to know what event occurred, the return value of the original call,
// and the metadata that is being carried over to this handler.
type message struct {
	// Key is the key/topic that this message is being published to.
	Key string
	// Route contains useful information about the service method invocation that triggered this publish.
	Route metadata.EndpointRoute
	// Metadata represents the encoded version of all metadata attributes stored on
	// the context that we want to follow the caller as it goes from service to service.
	Metadata metadata.EncodedBytes
	// Values is the return value of the service method that just completed. It will be passed
	// as the input of the subscriber(s) when they handle this event. It's the flattened value
	// map representation of the response:
	//
	// Example:
	// {
	//   "ID": ["12345"],
	//   "Name": ["The Dude"],
	//   "ContactInfo.Email": ["dude@example.com"],
	//   "ContactInfo.PhoneNumber": ["123-456-7890"],
	//   "AuditTrail.Created": ["2022-11-11T18:48:25+00:00"],
	//   "AuditTrail.Modified": ["2022-11-11T18:55:43+00:00"],
	// }
	Values url.Values
	// ErrorStatus represents the HTTP-style status code of the failure. Will be 0 if the source call didn't fail.
	ErrorStatus int
	// ErrorMessage returns the "err.Error()" value of the failure (if there was one). Will be "" if the call didn't fail.
	ErrorMessage string
}

// ErrorHandler returns true if this published message represents a method call that failed and is being routed to
// something like "FooService.Bar:Error" rather than "FooService.Bar".
func (m message) ErrorHandler() bool {
	return strings.HasSuffix(m.Key, errorKeySuffix)
}

// publishMiddleware defines the unit of work that every service endpoint should perform to publish
// their "I just finished this service function" event; the thing that drives our event gateway.
func publishMiddleware(broker eventsource.Broker, encoder codec.Encoder, valueEncoder codec.ValueEncoder, errorListener ErrorListener) services.MiddlewareFunc {
	return func(ctx context.Context, req any, next services.HandlerFunc) (any, error) {
		response, err := next(ctx, req)

		// We want the successful invocation to be propagated back to the caller as quickly
		// as possible, so don't wait for event publishing to happen in order to do that. This
		// does mean, however, that we need to perform asynchronous error handling w/ callbacks.
		// Even if we screw up the publishing portion, we still want the successful result to
		// make it back to the original caller.
		go func() {
			encodedMetadata := metadata.Encode(ctx)
			endpoint := metadata.Route(ctx)

			// We need a context separate from the overall request context. The original one
			// is likely some HTTP request context that will be canceled in a matter of
			// milliseconds because we'll have responded to the original call already. We don't
			// want our publish call to fail even if it wants to fire a nanosecond after the
			// request is done.
			pubCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // make configurable?
			defer cancel()

			msg := message{
				Route:    endpoint,
				Metadata: encodedMetadata,
			}

			switch {
			case err == nil:
				msg.Key = endpoint.QualifiedName()
				msg.Values = valueEncoder.EncodeValues(response)
			case err != nil:
				msg.Key = endpoint.QualifiedName() + errorKeySuffix
				msg.Values = valueEncoder.EncodeValues(req)
				msg.ErrorStatus = fail.Status(err)
				msg.ErrorMessage = err.Error()
			}

			buf := &bytes.Buffer{}
			if err = encoder.Encode(buf, msg); err != nil {
				errorListener(endpoint, err)
				return
			}
			if err = broker.Publish(pubCtx, msg.Key, buf.Bytes()); err != nil {
				errorListener(endpoint, err)
				return
			}
		}()
		return response, err
	}
}

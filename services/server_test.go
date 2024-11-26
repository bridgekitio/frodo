//go:build integration

package services_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/bridgekit-io/frodo/internal/quiet"
	"github.com/bridgekit-io/frodo/internal/testext"
	gen "github.com/bridgekit-io/frodo/internal/testext/gen"
	"github.com/bridgekit-io/frodo/services"
	"github.com/bridgekit-io/frodo/services/gateways/apis"
	"github.com/bridgekit-io/frodo/services/gateways/events"
	"github.com/stretchr/testify/suite"
)

func TestServerSuite(t *testing.T) {
	suite.Run(t, &ServerSuite{addresses: testext.NewFreeAddress("localhost", 20000)})
}

type ServerSuite struct {
	suite.Suite
	addresses      testext.FreeAddress
	httpMiddleware apis.HTTPMiddlewareFuncs
	httpClient     *http.Client
	httpAddress    string
	client         testext.SampleService
}

func (suite *ServerSuite) SetupTest() {
	suite.httpMiddleware = apis.HTTPMiddlewareFuncs{}
	suite.httpClient = &http.Client{Timeout: 2 * time.Second}
	suite.client = nil
}

func (suite *ServerSuite) start() (*services.Server, *testext.Sequence, func()) {
	// Grab a fresh address, so we can parallelize our tests.
	suite.httpAddress = suite.addresses.Next()
	suite.client = gen.SampleServiceClient(suite.httpAddress)

	// Capture invocations across both services in one timeline.
	sequence := &testext.Sequence{}

	sampleService := testext.SampleServiceHandler{Sequence: sequence}
	otherService := testext.OtherServiceHandler{
		Sequence:      sequence,
		SampleService: gen.SampleServiceClient(suite.httpAddress),
	}

	server := services.NewServer(
		services.Listen(apis.NewGateway(suite.httpAddress, apis.WithMiddleware(suite.httpMiddleware...))),
		services.Listen(events.NewGateway()),
		services.Register(gen.SampleServiceServer(sampleService)),
		services.Register(gen.OtherServiceServer(otherService)),
		services.OnPanic(func(err error, stack []byte) {
			sequence.Append("OnPanic:" + err.Error())
		}),
	)
	go func() { _ = server.Run(context.Background()) }()

	// Kinda crappy, but we need some time to make sure the server is up. Sometimes
	// this goes so fast that the test case fires before the server is fully running.
	// As a result the cases fail because the server's not running... duh.
	time.Sleep(25 * time.Millisecond)

	return server, sequence, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}
}

func (suite *ServerSuite) responseText(res any) string {
	if sampleRes, ok := res.(*testext.SampleResponse); ok {
		return sampleRes.Text
	}
	if otherRes, ok := res.(*testext.OtherResponse); ok {
		return otherRes.Text
	}
	if failRes, ok := res.(*testext.FailAlwaysResponse); ok {
		return failRes.ResponseValue
	}
	return "<invalid response type>"
}

func (suite *ServerSuite) responseStream(res any) *services.StreamResponse {
	if res == nil {
		return nil
	}
	if sampleRes, ok := res.(*testext.SampleDownloadResponse); ok {
		return &sampleRes.StreamResponse
	}
	if sampleRes, ok := res.(*testext.SampleRedirectResponse); ok {
		return &sampleRes.StreamResponse
	}
	return nil
}

func (suite *ServerSuite) streamContent(stream *services.StreamResponse) string {
	if stream == nil {
		return ""
	}
	content := stream.Content()
	if content == nil {
		return ""
	}

	defer quiet.Close(content)
	data, err := io.ReadAll(content)
	if err != nil {
		return ""
	}
	return string(data)
}

func (suite *ServerSuite) assertInvoked(calls *testext.Sequence, expected []string) {
	time.Sleep(50 * time.Millisecond)
	suite.ElementsMatch(calls.Values(), expected)
}

func (suite *ServerSuite) TestBasicExecution() {
	server, _, shutdown := suite.start()
	defer shutdown()

	res, err := server.Invoke(context.Background(), "SampleService", "Defaults", &testext.SampleRequest{
		Text: "Hello",
	})
	suite.Require().NoError(err)
	suite.Equal("Defaults:Hello", suite.responseText(res))
}

func (suite *ServerSuite) TestStreamedResponse() {
	server, _, shutdown := suite.start()
	defer shutdown()

	res, err := server.Invoke(context.Background(), "SampleService", "Download", &testext.SampleDownloadRequest{
		Format: "text/plain",
	})
	suite.Require().NoError(err)
	stream := suite.responseStream(res)
	suite.Equal("Donny, you're out of your element!", suite.streamContent(stream))
	suite.Equal("text/plain", stream.ContentType())
	suite.Equal("dude.txt", stream.ContentFileName())
	suite.Equal(34, stream.ContentLength())
}

func (suite *ServerSuite) TestResumableStreamedResponse() {
	server, _, shutdown := suite.start()
	defer shutdown()

	res, err := server.Invoke(context.Background(), "SampleService", "DownloadResumable", &testext.SampleDownloadRequest{
		Format: "text/plain",
	})
	suite.Require().NoError(err)
	stream := suite.responseStream(res)
	suite.Equal("<h1>The Dude Abides</h1>", suite.streamContent(stream))
	suite.Equal("text/html", stream.ContentType())
	suite.Equal("dude.html", stream.ContentFileName())

	start, end, size := stream.ContentRange()
	suite.Equal(50, start)
	suite.Equal(74, end)
	suite.Equal(1024, size)
}

// Ensure that the service still manages an endpoint even if it's not exposed by any gateway.
func (suite *ServerSuite) TestOmittedEndpoint() {
	server, _, shutdown := suite.start()
	defer shutdown()

	res, err := server.Invoke(context.Background(), "SampleService", "OmitMe", &testext.SampleRequest{
		Text: "Hello",
	})
	suite.Require().NoError(err)
	suite.Equal("Doesn't matter...", suite.responseText(res))
}

// Make sure that regardless of how endpoints are invoked, they trigger gateway
// specific middleware such as event publishing.
func (suite *ServerSuite) TestEvents() {
	server, calls, shutdown := suite.start()
	defer shutdown()

	calls.Reset()
	res, err := server.Invoke(context.Background(), "SampleService", "TriggerLowerCase", &testext.SampleRequest{Text: "Abide"})
	suite.Require().NoError(err)
	suite.Equal("abide", suite.responseText(res))
	suite.assertInvoked(calls, []string{
		"TriggerLowerCase:Abide",
		"ListenerB:abide", // event should receive the result of TriggerLowerCase
	})

	calls.Reset()
	res, err = server.Invoke(context.Background(), "SampleService", "TriggerError", &testext.SampleRequest{Text: "Abide"})
	suite.Require().Error(err)
	suite.assertInvoked(calls, []string{})
}

// Ensure that ServiceA is able to listen to events from ServiceB and that ServiceB
// can listen to events from ServiceA as well. As a side effect, this one also
// makes sure that event triggers and cascade and cause others to trigger.
func (suite *ServerSuite) TestEvents_bidirectional() {
	server, calls, shutdown := suite.start()
	defer shutdown()

	calls.Reset()
	sampleRes, err := server.Invoke(context.Background(), "SampleService", "TriggerUpperCase", &testext.SampleRequest{Text: "Abide"})
	suite.Require().NoError(err)
	suite.Equal("ABIDE", suite.responseText(sampleRes))
	suite.assertInvoked(calls, []string{
		"TriggerUpperCase:Abide",
		"ListenerA:ABIDE",
		"ListenerB:ABIDE",
		"ListenerB:ListenerA:ABIDE", // Cascade when ListenerB fired after ListenerA fired
		"ListenWell:ABIDE",          // OtherService also gets SampleService.TriggerUpperCase events
	})

	calls.Reset()
	otherRes, err := server.Invoke(context.Background(), "OtherService", "SpaceOut", &testext.OtherRequest{Text: "Abide"})
	suite.Require().NoError(err)
	suite.Equal("A b i d e", suite.responseText(otherRes))
	suite.assertInvoked(calls, []string{
		"SpaceOut:Abide",       // The original call we invoked.
		"ListenerB:A b i d e",  // The subscriber in the SampleService
		"ListenWell:A b i d e", // The subscriber in the OtherService
	})
}

// Ensure that one service can invoke functions on another and that those functions can cause
// events to fire and be handled on multiple services.
func (suite *ServerSuite) TestRPCWithEvents() {
	server, calls, shutdown := suite.start()
	defer shutdown()

	calls.Reset()
	res, err := server.Invoke(context.Background(), "OtherService", "RPCExample", &testext.OtherRequest{Text: "Abide"})
	suite.Require().NoError(err)
	suite.Equal("ABIDE", suite.responseText(res))
	suite.assertInvoked(calls, []string{
		// Our initial call a few lines above.
		"RPCExample:Abide",

		// The RPC call that the other service made to the sample service.
		"TriggerUpperCase:Abide",

		// Since RPCExample explicitly calls TriggerUpperCase, when *that* call finishes, it
		// should trigger events that are handled by methods on both services.
		"ListenerA:ABIDE",
		"ListenerB:ABIDE",
		"ListenerB:ListenerA:ABIDE", // Cascade when ListenerB fired after ListenerA fired
		"ListenWell:ABIDE",
	})
}

// Ensures that you can invoke a method which triggers the event gateway to run another method. Then,
// when THAT method finishes, it triggers even more event-based methods. It ensures that we properly
// support more complex event flows.
func (suite *ServerSuite) TestEventChain() {
	server, calls, shutdown := suite.start()
	defer shutdown()

	calls.Reset()
	res, err := server.Invoke(context.Background(), "OtherService", "ChainOne", &testext.OtherRequest{Text: "Abide"})
	suite.Require().NoError(err)
	suite.Equal("ChainOne:Abide", suite.responseText(res))
	suite.assertInvoked(calls, []string{
		// Our initial call a few lines above.
		"ChainOne:Abide",

		// The first call that is triggered by the success of ChainOne. The failure one does run and append to
		// the sequence, but it does NOT trigger ChainFailAfter.
		"ChainTwo:ChainOne:Abide",
		"ChainFail:ChainOne:Abide",

		// When ChainTwo completes, these are invoked.
		"ChainThree:ChainTwo:ChainOne:Abide",
		"ChainFour:ChainTwo:ChainOne:Abide",
	})
}

func (suite *ServerSuite) TestPanic() {
	server, calls, shutdown := suite.start()
	defer shutdown()

	calls.Reset()
	_, err := server.Invoke(context.Background(), "SampleService", "Panic", &testext.SampleRequest{Text: "Abide"})

	// First, make sure that the message in the panic is ultimately returned as the error message from the call.
	// The handler just calls panic("don't"), so "don't" should be the returned error message.
	suite.Require().Error(err)
	suite.Require().Equal("don't", err.Error())

	// We also want to make sure that our OnPanic() callback was properly invoked. This ensures that our panic
	// logging hook is properly working. This is not added to the calls sequence in the service, but rather in
	// suite.start() when setting up the service.
	suite.assertInvoked(calls, []string{"OnPanic:don't"})
}

// Prevent regression on the bug where HTTP middleware would fire twice for every single call.
// https://github.com/bridgekit-io/frodo/issues/2
func (suite *ServerSuite) TestHttpMiddlewareFireOnce() {
	middlewareSequence := &testext.Sequence{}

	suite.httpMiddleware = apis.HTTPMiddlewareFuncs{
		func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
			middlewareSequence.Append("1:BEFORE")
			next(w, req)
			middlewareSequence.Append("1:AFTER")
		},
		func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
			middlewareSequence.Append("2:BEFORE")
			next(w, req)
			middlewareSequence.Append("2:AFTER")
		},
	}

	_, _, shutdown := suite.start()
	defer shutdown()

	res, err := suite.client.Defaults(context.Background(), &testext.SampleRequest{Text: "Hello"})

	suite.Require().NoError(err)
	suite.Equal("Defaults:Hello", suite.responseText(res))
	suite.Equal([]string{"1:BEFORE", "2:BEFORE", "2:AFTER", "1:AFTER"}, middlewareSequence.Values())
}

// Ensures that you can invoke a method which triggers the event gateway to run another method when some other
// service method FAILS (i.e. returns a non-nil error).
func (suite *ServerSuite) TestEventErrorChain() {
	server, calls, shutdown := suite.start()
	defer shutdown()

	calls.Reset()
	res, err := server.Invoke(context.Background(), "SampleService", "FailAlways", &testext.FailAlwaysRequest{RequestValue: "Abide"})
	suite.Require().Error(err)
	suite.Equal("Do Not Abide", suite.responseText(res)) // the result of the failed call - it DOES return a value

	// Some key aspects of what we're testing here. First, FailAlwaysErrorRequest has attributes that match both the
	// request AND response of the method that is supposed to fail.
	//
	// (A) We expect that, on failure, the gateway will send the REQUEST data from the failed method call, not the
	// response data. After all, the call failed so there probably isn't any valid response data to send over. By
	// sending the request data, it gives the error handling subscriber the chance to look at the inputs of the
	// failed call and do something intelligent with it.
	//
	// (B) The error that occurred will be marshalled to the subscriber and bound to various "Error.XXX" fields. We
	// support binding to all manner of name to give the same flexibility that we do in the 'fail' package.
	suite.assertInvoked(calls, []string{
		// Our initial call a few lines above.
		"FailAlways",

		// The :Error handler
		"OnFailAlways.Request:Abide",
		"OnFailAlways.Response:",
		"OnFailAlways.Error.Error:a world of pain",
		"OnFailAlways.Error.Message:a world of pain",
		"OnFailAlways.Error.Code:501",
		"OnFailAlways.Error.Status:501",
		"OnFailAlways.Error.StatusCode:501",
		"OnFailAlways.Error.HTTPStatusCode:501",
	})
}

// Ensures that if you have a multi-stage event chain of service calls (e.g. CallA->CallB->CallC), the chain is stopped
// if an error occurs in the middle of the chain. For instance, if "CallB" fails, then we should see the execution of
// the "ON Service.CallB:Error" instead of "ON Service.CallC".
func (suite *ServerSuite) TestEventChainErrorMidSequence() {
	server, calls, shutdown := suite.start()
	defer shutdown()

	calls.Reset()
	res, err := server.Invoke(context.Background(), "SampleService", "Chain1", &testext.SampleRequest{Text: "Abide"})
	suite.Require().NoError(err)
	suite.Require().Equal("Abide", suite.responseText(res))

	suite.assertInvoked(calls, []string{
		// The first call happens normally, as  do the calls that are not part of the standard, shared consumer group.
		// NOTE: we can only rely on this ordering because we're using the local broker for these tests.
		"Chain1:Abide",
		"Chain1GroupStar:*:Abide",
		"Chain1GroupFooBar:FooBar:Abide",

		// The second call does fire, it just returns an error.
		"Chain2:Abide",

		// The third call should be "Chain2OnError", not the OnSuccess version.
		//
		// IMPORTANT DETAIL: The 'Text' attribute is still "Abide" because the error event receives the INPUT/REQUEST
		// of the call that failed. The non-error return value of Chain2 is ignored even though it's non-null with a
		// different value for 'Text'. Since the error input is driven off of Chain2's request, it should still be Abide.
		"Chain2OnError.Text:Abide",
		"Chain2OnError.Error.Error:this will not stand",
		"Chain2OnError.Error.Message:this will not stand",
		"Chain2OnError.Error.Code:404",
		"Chain2OnError.Error.Status:404",
		"Chain2OnError.Error.StatusCode:404",
		"Chain2OnError.Error.HTTPStatusCode:404",
	})
}

func (suite *ServerSuite) TestRestoreAuthorizationMiddleware() {
	_, values, shutdown := suite.start()
	defer shutdown()

	Authorization := "Authorization"
	SecWebsocketProtocol := "Sec-WebSocket-Protocol"

	testCase := func(expected string, updateHeaders func(req *http.Request)) {
		req, _ := http.NewRequest("POST", "http://"+suite.httpAddress+"/v2/SampleService.Authorization", nil)
		updateHeaders(req)
		suite.assertRequestSuccess(suite.httpClient.Do(req))
		suite.Require().Equal("Authorization:"+expected, values.Last())
	}

	// Don't apply any headers.
	testCase("", func(req *http.Request) {
	})

	// Explicitly blank
	testCase("", func(req *http.Request) {
		req.Header.Set(Authorization, "")
	})

	// Just a token value
	testCase("123", func(req *http.Request) {
		req.Header.Set(Authorization, "123")
	})

	// A scheme/value separated value
	testCase("Bearer 456", func(req *http.Request) {
		req.Header.Set(Authorization, "Bearer 456")
	})

	// Seriously, frodo doesn't care about the auth format - that's a concern for your app
	testCase("The beer has gone bad", func(req *http.Request) {
		req.Header.Set(Authorization, "The beer has gone bad")
	})

	// Websocket auth needs to conform to the single-token format like "Authorization.SCHEME.VALUE" or "Authorization.VALUE", so this won't work
	testCase("", func(req *http.Request) {
		req.Header.Set(SecWebsocketProtocol, "Authorization: Bearer 12345")
	})
	testCase("", func(req *http.Request) {
		req.Header.Set(SecWebsocketProtocol, "Authorization-Bearer-12345")
	})

	// Can fall back to using Sec-Websocket-Protocol when necessary (value only)
	testCase("12345", func(req *http.Request) {
		req.Header.Set(SecWebsocketProtocol, "Authorization.12345")
	})

	// Splits common schemes like "Basic-123" into the more canonical "Basic 123"
	testCase("Basic ABC", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.Basic.ABC") })
	testCase("Bearer ABC", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.Bearer.ABC") })
	testCase("Digest ABC", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.Digest.ABC") })
	testCase("Token ABC", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.Token.ABC") })
	testCase("HOBA ABC", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.HOBA.ABC") })
	testCase("Mutual ABC", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.Mutual.ABC") })
	testCase("VAPID ABC", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.VAPID.ABC") })
	testCase("SCRAM ABC", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.SCRAM.ABC") })
	testCase("AWS4-HMAC-SHA256 ABC", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.AWS4-HMAC-SHA256.ABC") })

	// Obscure/unknown schemes are left w/ the "-" splitting them. Sorry.
	testCase("FooBar.ABC", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.FooBar.ABC") })

	// Our scheme checks are CASE SENSITIVE... follow the standards, you scallywag, otherwise we're leaving the "-" in between them.
	testCase("BaSIc.ABC", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.BaSIc.ABC") })
	testCase("bearer.ABC", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.bearer.ABC") })

	// You should be allowed to have periods in your token value.
	testCase("Bearer ABC.123..XYZ.", func(req *http.Request) { req.Header.Set(SecWebsocketProtocol, "Authorization.Bearer.ABC.123..XYZ.") })

	// If you provide both, standard Authorization header wins.
	testCase("Bearer 123", func(req *http.Request) {
		req.Header.Set(Authorization, "Bearer 123")
		req.Header.Set(SecWebsocketProtocol, "Authorization.Basic.456")
	})
}

func (suite *ServerSuite) assertRequestSuccess(res *http.Response, err error) {
	suite.Require().NoError(err, "HTTP request should have a low level 'err' failure.")
	suite.Require().True(res.StatusCode < 400, "HTTP request should have a successful status code")
}

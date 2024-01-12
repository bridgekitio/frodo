package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bridgekitio/frodo/codec"
	"github.com/bridgekitio/frodo/fail"
	"github.com/bridgekitio/frodo/internal/naming"
	"github.com/bridgekitio/frodo/internal/quiet"
	"github.com/bridgekitio/frodo/services"
)

// NewClient constructs the RPC client that does the "heavy lifting" when communicating
// with remote abide-powered RPC services. It contains all data/logic required to marshal/unmarshal
// requests/responses as well as communicate w/ the remote service.
func NewClient(name string, addr string, options ...ClientOption) Client {
	// Allow either: "foo:8080" or "http://foo:8080" or "https://foo:8080"
	switch {
	case addr == "":
	case strings.HasPrefix(addr, "http://"):
	case strings.HasPrefix(addr, "https://"):
	default:
		// Since you didn't provide a protocol, so default to plain HTTP
		addr = "http://" + strings.TrimPrefix(addr, "/")
	}

	defaultTimeout := 30 * time.Second
	dialer := &net.Dialer{Timeout: defaultTimeout}
	client := Client{
		HTTP: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				DialContext:         dialer.DialContext,
				TLSHandshakeTimeout: defaultTimeout,
			},
		},
		Name:       name,
		BaseURL:    strings.TrimSuffix(addr, "/"),
		codecs:     codec.New(),
		middleware: clientMiddlewarePipeline{},
	}
	for _, option := range options {
		option(&client)
	}

	// Let the user's custom middleware do whatever the hell it wants to the context/request
	// before our standard middleware finalizes everything.
	client.middleware = append(client.middleware,
		writeMetadataHeader,
		writeAuthorizationHeader,
	)
	client.roundTrip = client.middleware.Then(client.HTTP.Do)
	return client
}

// ClientOption is a single configurable setting that modifies some attribute of the RPC client
// when building one via NewClient().
type ClientOption func(*Client)

// Client manages all RPC communication with other abide-powered services. It uses HTTP under the hood,
// so you can supply a custom HTTP client by including WithHTTPClient() when calling your client
// constructor, NewXxxServiceClient().
type Client struct {
	// HTTP takes care of the raw HTTP request/response logic used when communicating w/ remote services.
	HTTP *http.Client
	// BaseURL contains the protocol/host/port/etc that is the prefix for all service function
	// endpoints. (e.g. "http://api.myawesomeapp.com")
	BaseURL string
	// Name is just the display name of the service; used only for debugging/tracing purposes.
	Name string
	// Codes maintains decoders we can use to read in different types of response bodies.
	codecs codec.Registry
	// Middleware defines all of the units of work we will apply to the request/response when
	// round-tripping our RPC call to the remote service.
	middleware clientMiddlewarePipeline
	// roundTrip captures all middleware and the actual request dispatching in a single handler
	// function. This is what we'll call once we've created the HTTP/RPC request when invoking
	// one of your client's service functions.
	roundTrip RoundTripperFunc
}

// Invoke handles the standard request/response logic used to call a service method on the remote service.
// You should NOT call this yourself. Instead, you should stick to the strongly typed, code-generated
// service functions on your client.
func (c Client) Invoke(ctx context.Context, method string, path string, serviceRequest any, serviceResponse any) error {
	// Step 1: Fill in the URL path and query string w/ fields from the request. (e.g. /user/{id} -> /user/abc)
	// If this is a GET/DELETE/etc. that doesn't support bodies, this will include a query string
	// with the remaining service request values.
	address := c.buildURL(method, path, serviceRequest)

	// Step 2: Create a JSON reader for the request body (POST/PUT/PATCH only).
	body, err := c.createRequestBody(method, serviceRequest)
	if err != nil {
		return fmt.Errorf("unable to create request body: %w", err)
	}

	// Step 3: Form the HTTP request
	request, err := http.NewRequestWithContext(ctx, method, address, body)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	// Step 4: Run the request through all middleware and fire it off.
	response, err := c.roundTrip(request)
	if err != nil {
		return fmt.Errorf("round trip error: %w", err)
	}

	// Step 5: Based on the status code, either populate "out" struct (service response) with the
	// decoded body/JSON or respond a properly formed error.
	err = c.decodeResponse(response, serviceResponse)
	if err != nil {
		return fmt.Errorf("unable to decode response: %w", err)
	}
	return nil
}

func (c Client) decodeResponse(response *http.Response, serviceResponse any) error {
	if response.StatusCode >= 400 {
		return c.decodeError(response)
	}
	if raw, ok := serviceResponse.(services.ContentGetter); ok {
		return c.decodeResponseStream(response, raw)
	}
	return c.decodeResponseValue(response, serviceResponse)
}

func (c Client) decodeResponseValue(res *http.Response, serviceResponse any) error {
	defer quiet.Close(res.Body)

	contentType := res.Header.Get("Content-Type")
	decoder := c.codecs.Decoder(contentType)

	if err := decoder.Decode(res.Body, serviceResponse); err != nil {
		return fmt.Errorf("rpc: unable to decode response: %w", err)
	}
	return nil
}

func (c Client) decodeResponseStream(res *http.Response, streamResponse services.ContentGetter) error {
	switch setter, ok := streamResponse.(services.ContentSetter); ok {
	case true:
		// We do NOT auto-close the body because we have no idea what you plan to do
		// with the stream. It may be much bigger than we want to keep in memory, so
		// we don't want to just copy it to a bytes.Buffer{} and close res.Body. Since
		// the user will consume the stream long after the request scope is done, we
		// need to rely on the user to close it when they're done.
		setter.SetContent(res.Body)
	case false:
		// We're not able to apply the body to the service response struct. Since
		// we have no way to actually consume the body, just close it so that we
		// don't end up leaking connections.
		quiet.Close(res.Body)
	}

	// Based on whether the response type implements any of the other ContentXxxSetter
	// interfaces, we'll apply more of the additional stream metadata back onto the
	// response value we're trying to populate.

	if setter, ok := streamResponse.(services.ContentTypeSetter); ok {
		setter.SetContentType(res.Header.Get("Content-Type"))
	}

	if setter, ok := streamResponse.(services.ContentLengthSetter); ok {
		length, _ := strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
		setter.SetContentLength(int(length))
	}

	if setter, ok := streamResponse.(services.ContentRangeSetter); ok {
		// Separate "bytes 0-100/1000" into "bytes" and "0-100/1000". For now, we
		// don't allow anything other than "bytes", so we're just ignoring that value.
		_, rangeValue, _ := strings.Cut(res.Header.Get("Content-Range"), " ")

		startEndValue, sizeValue, _ := strings.Cut(rangeValue, "/")
		startValue, endValue, _ := strings.Cut(startEndValue, "-")

		start, _ := strconv.ParseInt(strings.TrimSpace(startValue), 10, 64)
		end, _ := strconv.ParseInt(strings.TrimSpace(endValue), 10, 64)
		size, _ := strconv.ParseInt(strings.TrimSpace(sizeValue), 10, 64) // "*" will end up 0 anyway

		setter.SetContentRange(int(start), int(end), int(size))
	}

	if setter, ok := streamResponse.(services.ContentFileNameSetter); ok {
		fileName := naming.DispositionFileName(res.Header.Get("Content-Disposition"))
		setter.SetContentFileName(fileName)
	}

	return nil
}

// newStatusError takes the response (assumed to be a 400+ status already) and creates
// an RPCError with the proper HTTP status as it tries to preserve the original error's message.
func (c Client) decodeError(r *http.Response) error {
	defer quiet.Close(r.Body)

	errData, _ := io.ReadAll(r.Body)
	contentType := r.Header.Get("Content-Type")

	// If the server didn't return JSON, assume that it's just plain text w/ the message to propagate
	// as you'd get if you invoked `http.Error()`
	if !strings.HasPrefix(contentType, "application/json") {
		return fail.New(r.StatusCode, "rpc: %s", string(errData))
	}

	// As JSON, it's likely that the JSON is one of these formats:
	//
	// "Just the message"
	//    or
	// {"status":404, "message": "not found, dummy"}
	//
	// Based on what it looks like, unmarshal accordingly.
	if strings.HasPrefix(string(errData), `"`) && strings.HasSuffix(string(errData), `"`) {
		err := ""
		_ = json.Unmarshal(errData, &err)
		return fail.New(r.StatusCode, "rpc error: %s", err)
	}
	if strings.HasPrefix(string(errData), `{`) {
		err := fail.StatusError{}
		_ = json.Unmarshal(errData, &err)
		return fail.New(r.StatusCode, "rpc error: %s", err.Error())
	}

	// It's JSON, but it's a format we don't recognize, so no message for you. Keep the status, though.
	return fail.New(r.StatusCode, "service invocation error")
}

func (c Client) createRequestBody(method string, serviceRequest any) (io.Reader, error) {
	switch method {
	case http.MethodPut, http.MethodPost, http.MethodPatch:
		body := &bytes.Buffer{}
		err := c.codecs.DefaultEncoder().Encode(body, serviceRequest)
		return body, err
	default:
		return nil, nil
	}
}

func (c Client) buildURL(method string, path string, serviceRequest any) string {
	attributes := c.codecs.DefaultValueEncoder().EncodeValues(serviceRequest)

	path = strings.Trim(path, "/")
	pathSegments := naming.TokenizePath(path, '/')

	// Using the mapping of field names to request values (attributes), fill in the path
	// pattern with real value request values.
	//
	// Example: "/user/{UserID}/message/{ID}" --> "/user/1234/message/5678"
	for i, pathSegment := range pathSegments {
		// Leave fixed segments alone (e.g. "user" in "/user/{id}/messages", but not "{id}")
		if !naming.IsPathVariable(pathSegment) {
			continue
		}

		// Replace path param variables w/ the equivalent value from the service request. Make
		// sure to path escape the values. For instance if we're filling in values for the
		// pattern "/content-type/{ContentType}" and the value for "{ContentType}" is "image/png"
		// we want the final URL to be "/content-type/image%2Fpng" and not "/content-type/image/png"
		// because you'd be sneaking in more path segments.
		paramName := pathSegment[1 : len(pathSegment)-1]
		pathSegments[i] = url.PathEscape(attributes.Get(paramName))

		// Remove the attribute, so it doesn't also get encoded in the query string, also.
		attributes.Del(paramName)
	}

	address := c.BaseURL + "/" + strings.Join(pathSegments, "/")
	switch method {
	case http.MethodPut, http.MethodPost, http.MethodPatch:
		// If we're doing a POST/PUT/PATCH, don't bother adding query string arguments. Non-path
		// values will just be part of the JSON structure in the request's body.
		return address
	default:
		// We're doing a GET/DELETE/etc, so all request values must come via query string args.
		return address + "?" + attributes.Encode()
	}
}

// fixedSegment returns true if the given URL path segment is not wrapped in "{}" indicating that it's a variable.
func (c Client) fixedSegment(segment string) bool {
	return !strings.HasPrefix(segment, "{") || !strings.HasSuffix(segment, "}")
}

// WithMiddleware sets the chain of HTTP request/response handlers you want to invoke
// on each service function invocation before/after we dispatch the HTTP request.
func WithMiddleware(funcs ...ClientMiddlewareFunc) ClientOption {
	return func(client *Client) {
		client.middleware = funcs
	}
}

// WithHTTPClient allows you to provide an HTTP client configured to your liking. You do not *need*
// to supply this. The default client already implements a 30 second timeout, but if you want a
// different timeout or custom dialer/transport/etc, then you can feed in you custom client here and
// we'll use that one for all HTTP communication with other services.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(rpcClient *Client) {
		rpcClient.HTTP = httpClient
	}
}

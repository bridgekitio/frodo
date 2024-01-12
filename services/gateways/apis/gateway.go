package apis

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/bridgekitio/frodo/codec"
	"github.com/bridgekitio/frodo/fail"
	"github.com/bridgekitio/frodo/internal/naming"
	"github.com/bridgekitio/frodo/internal/quiet"
	"github.com/bridgekitio/frodo/services"
	"github.com/dimfeld/httptreemux/v5"
)

// NewGateway creates a new API Gateway that allows your service to accept incoming requests
// using RPC over HTTP. This encapsulates a standard net/http server while providing options
// so that you can customize various aspects of the server, TLS, and middleware as desired.
func NewGateway(address string, options ...GatewayOption) *Gateway {
	router := httptreemux.New()
	gw := Gateway{
		router:     router,
		codecs:     codec.New(),
		middleware: HTTPMiddlewareFuncs{},
		endpoints:  map[httpRoute]services.Endpoint{},
		server:     &http.Server{Addr: address, Handler: router},
		tlsCert:    "",
		tlsKey:     "",
	}
	for _, option := range options {
		option(&gw)
	}

	gw.router.NotFoundHandler = gw.notFoundHandler
	return &gw
}

// Gateway encapsulates all of the HTTP(S) server and routing components that enable your
// service to accept incoming requests over RPC/HTTP.
//
// DO NOT CREATE THIS DIRECTLY. Use the NewGateway() constructor to properly set up an
// API gateway in your main() function.
type Gateway struct {
	codecs     codec.Registry
	middleware HTTPMiddlewareFuncs
	endpoints  map[httpRoute]services.Endpoint
	router     *httptreemux.TreeMux
	server     *http.Server
	tlsCert    string
	tlsKey     string
}

// Type returns "API" to properly tag this type of gateway.
func (gw *Gateway) Type() services.GatewayType {
	return services.GatewayTypeAPI
}

// Listen fires up the underlying HTTP web server and blocks just like the net/http
// web server code already does. The only difference is that when the gateway shuts
// down gracefully, this will return nil instead of http.ErrServerClosed. All other
// errors are propagated back.
func (gw *Gateway) Listen() error {
	switch err := gw.listenAndServe(); err {
	case nil, http.ErrServerClosed:
		return nil
	default:
		return fmt.Errorf("api gateway error: %w", err)
	}
}

// listenAndServe determines if we need to start up in plain old HTTP mode or HTTPS
// using TLS certificates or a TLS config/manager you configured (i.e. lets encrypt).
// This will block until the server shuts down just like the underlying server does.
func (gw *Gateway) listenAndServe() error {
	switch {
	case gw.UseTLS():
		return gw.server.ListenAndServeTLS(gw.tlsCert, gw.tlsKey)
	default:
		return gw.server.ListenAndServe()
	}
}

// Shutdown attempts to gracefully shut down the HTTP server. It will wait for any in-progress
// requests to finish and then shut down (unblocking Listen()). You can provide a context
// with a deadline to limit how long you want to wait before giving up and shutting down anyway.
func (gw *Gateway) Shutdown(ctx context.Context) error {
	return gw.server.Shutdown(ctx)
}

// Register the operation with the gateway so that it can be exposed for invoking remotely.
func (gw *Gateway) Register(endpoint services.Endpoint, route services.EndpointRoute) {
	if route.GatewayType != services.GatewayTypeAPI {
		return
	}

	// The user specified a path like "GET /user/{id}" in their code, so when they fetch the
	// endpoint data later, that's what we want it to look like, so we'll leave the endpoint's
	// Path attribute alone. But... the router needs the full path which includes the optional
	// prefix (e.g. "/v2"). So we'll use the full path for routing and lookups (transparent to
	// the user), but the user will never have to see the "/v2" portion.
	//
	// And yes, we're ignoring the error, but that only happens if JoinPath can't parse the
	// first parameter as a URL. Since that's hardcoded to something that is guaranteed to parse
	// properly, we're good.
	path := gw.normalizePath(route)
	method := strings.ToUpper(route.Method)

	// We want to try and make sure that our bookkeeping tasks like request ids, metadata,
	// etc. are all done by the time any of the user's custom middleware or the handler fires.
	customFuncs := gw.middleware
	standardFuncs := HTTPMiddlewareFuncs{
		recoverFromPanic(gw.codecs.DefaultEncoder()),
		restoreMetadata(),
		restoreMetadataHeaders(),
		restoreMetadataEndpoint(endpoint, route),
		restoreTraceID(),
		restoreAuthorization(),
	}
	httpHandler := standardFuncs.Append(customFuncs...).Then(gw.toHTTPHandler(endpoint, route))

	// If you're registering "POST /FooService.Bar" we're going to create a route for
	// the POST as well as an additional, implicit OPTIONS route. This is so that
	// you can use WithMiddleware(Func) to enable CORS in your API. All of your middleware
	// is actually part of the router/mux handling (see comments in NewGateway() for details as to why), so
	// if we don't include an explicit OPTIONS route for this path then your CORS middleware
	// will never actually get invoked - the http router will just reject the request. We fully expect
	// your CORS middleware to short-circuit the 'next' chain, so the 405 failure we're hard-coding
	// as the OPTIONS handler won't actually be invoked if you enable CORS via middleware.
	gw.endpoints[httpRoute{Method: method, Path: path}] = endpoint
	gw.endpoints[httpRoute{Method: http.MethodOptions, Path: path}] = endpoint
	gw.router.UsingContext().Handle(method, path, gw.middleware.Then(httpHandler))
	gw.registerOptions(path)
}

// normalizePath takes all disparate pathing information at the service and endpoint levels and returns the
// actual path that we will feed to the underlying HTTP router. The steps include:
//
// - Converting /user/{User.ID} brace variable to  /user/:User.ID colon format.
func (gw *Gateway) normalizePath(route services.EndpointRoute) string {
	// The HTTP router expects variables to be in colon-format (e.g. ":User.ID"). We allow you to use
	// brace format (e.g. "{User.ID}"), so convert variables to the router's format.
	pathSegments := strings.Split(route.Path, "/")
	for i, segment := range pathSegments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			pathSegments[i] = ":" + segment[1:len(segment)-1]
		}
	}

	return strings.Join(pathSegments, "/")
}

func (gw *Gateway) registerOptions(path string) {
	// I realize that recovering from panics makes the baby jesus cry. This is to handle the case where you
	// register multiple service functions with the same path, but different methods. For instance:
	//
	//   GET  /foo/bar
	//   POST /foo/bar
	//
	// Since we blindly register an options with each, we will end up registering OPTIONS twice for that
	// path. The httptreemux will panic when that happens. Originally, I planned on just looking through the
	// gateway's already-registered endpoint paths for a match (and thus skip), but there's a case that's
	// hard to detect:
	//
	//   GET  /foo/{bar}
	//   POST /foo/{goo}
	//
	// A dumb string-based check would see those as unique paths, but the router will still barf because they
	// are functionally equivalent.
	//
	// So.... since the mux is already doing all of the hard work, I'm catching the panic in this
	// instance to make life easier. If there's something fundamentally wrong with the route, we'll fail
	// more naturally when we register the "real" endpoint route, so we're not going to miss meaningful errors.
	defer func() {
		recover()
	}()

	gw.router.UsingContext().OPTIONS(path, gw.middleware.Then(gw.methodNotAllowedHandler))
}

func (gw *Gateway) toHTTPHandler(endpoint services.Endpoint, route services.EndpointRoute) http.HandlerFunc {
	// For now, we're only going to support JSON for non-stream entities. At some point if we
	// want to support content negotiation, we can put this call inside the handler function
	// and use the "Accept" header to determine which codec we try to use.
	//
	// But for now... only JSON for you.
	encoder := gw.codecs.Encoder("application/json")
	decoder := gw.codecs.Decoder("application/json")
	valueDecoder := gw.codecs.ValueDecoder("application/json")

	return func(w http.ResponseWriter, req *http.Request) {
		// Create a blank request struct that we will populate w/ request body/path/query data.
		serviceRequest := endpoint.NewInput()

		// The order of these 3 layers of decoding is very intentional. It's pretty easy when your
		// request plays by the rules such as this:
		//
		//     PATCH /user/123?LastName=Lebowski
		//     {"Hobbies":["Bowling", "White Russians"]}
		//
		// We'll bind the user's ID to 123, the last name to "Lebowski", and the hobbies
		// to the 2-element array. It gets hairy when your request starts to look like this:
		//
		//     PATCH /user/123?LastName=Lebowski&ID=456&Hobbies=Driving&Hobbies=Bowling
		//     {"Hobbies":["Bowling", "White Russians"], "ID":"789"}
		//
		// What do you expect the ID to be since it's defined in 3 separate places? Similarly,
		// Hobbies is in the query string and body, so which is it?
		//
		// We treat these 3 value sources as a hierarchy where query string is the weakest of
		// values. We'll use those values if they're not defined elsewhere. Next is the body,
		// and the ultimate winner that can't be overridden is the path. The idea is that you'll
		// usually perform authorization and other tasks based on that value, so we don't want
		// you to authorize access to user 123 but then apply the update to 789 because that's
		// what was in the body.
		//
		// As a result, path params will always override anything in the
		// body or query string. The body will override anything defined in the query string. This
		// way you can't sneak in values to circumvent security while providing a sane set of
		// binding expectations to your input data.
		if err := valueDecoder.DecodeValues(req.URL.Query(), &serviceRequest); err != nil {
			respondFailure(w, req, encoder, err)
			return
		}
		if err := decoder.Decode(req.Body, &serviceRequest); err != nil {
			respondFailure(w, req, encoder, err)
			return
		}
		if err := valueDecoder.DecodeValues(pathParams(req), &serviceRequest); err != nil {
			respondFailure(w, req, encoder, err)
			return
		}

		serviceResponse, err := endpoint.Handler(req.Context(), serviceRequest)
		if err != nil {
			respondFailure(w, req, encoder, err)
			return
		}
		respondSuccess(w, req, encoder, serviceResponse, route.Status)
	}
}

// UseTLS returns false (default) when Listen() will fire up in normal HTTP mode. If
// this returns true then Listen() will fire up the underlying server in HTTPS mode
// using the TLS cert/config you provided when creating the Gateway.
func (gw *Gateway) UseTLS() bool {
	return gw.tlsCert != "" || gw.tlsKey != "" || gw.server.TLSConfig != nil
}

// methodNotAllowedHandler just replies with a 405 error status no matter what. It's the
// default OPTIONS handler we use so that you can insert the CORS middleware of your
// choice should you choose to enable browser-based communication w/ your service.
func (gw *Gateway) methodNotAllowedHandler(w http.ResponseWriter, req *http.Request) {
	encoder := gw.codecs.DefaultEncoder()
	respondFailure(w, req, encoder, fail.MethodNotAllowed("method not allowed: %v", req.Method))
}

// notFoundHandler replies with a 404 error status no matter what. The body will match our
// look like {"Status":404, "Message":"..."} to match our standard error payload.
func (gw *Gateway) notFoundHandler(w http.ResponseWriter, req *http.Request) {
	encoder := gw.codecs.DefaultEncoder()
	respondFailure(w, req, encoder, fail.NotFound("not found"))
}

// httpRoute is the key used to reference routes in the gateway's route table.
type httpRoute struct {
	// Method is the HTTP method used by this route (e.g. "GET", "POST", etc.).
	Method string
	// Path is the path pattern used by this route (e.g. "/user/{ID}")
	Path string
}

func respondFailure(w http.ResponseWriter, _ *http.Request, encoder codec.Encoder, err error) {
	status := fail.Status(err)
	w.Header().Set("Content-Type", encoder.ContentType())
	w.WriteHeader(status)
	_ = encoder.Encode(w, fail.New(status, err.Error()))
}

func respondSuccess(w http.ResponseWriter, req *http.Request, encoder codec.Encoder, serviceResponse any, status int) {
	// If your response implements either of the redirect getter methods, try to forward on to
	// the desired address using either a 307/308 as needed.
	//
	// IMPORTANT: Check for redirects BEFORE checking for content stream responses. It's possible
	// that your response struct implements both Set/Redirect() and Set/Content(). This allows you
	// to redirect to some raw content like a file on S3 or something like that. The gateway will
	// follow the redirect to your signed URL. The client can then fill in the file details
	// on the caller's response instance. Nifty!
	redirect, ok := serviceResponse.(services.Redirector)
	if ok && respondSuccessRedirect(w, req, redirect) {
		return
	}
	redirectPermanent, ok := serviceResponse.(services.RedirectorPermanent)
	if ok && respondSuccessRedirectPermanent(w, req, redirectPermanent) {
		return
	}

	// The method's response appears to want to send raw bytes itself rather than relying
	// on the auto-JSON (or whatever encoding) that we normally use to marshal responses.
	// Based on the methods implemented by the response struct, we can send a response w/ different
	// headers in addition to the raw bytes. See the docs for RespondRawRanged, RespondRawSized,
	// and RespondRaw for more info on what headers we'll include.
	streamResponse, ok := serviceResponse.(services.ContentGetter)
	if ok && respondSuccessStream(w, streamResponse, status) {
		return
	}

	// Just encode the response struct/value and deliver it to the caller.
	w.Header().Set("Content-Type", encoder.ContentType())
	w.WriteHeader(status)
	_ = encoder.Encode(w, serviceResponse)
}

func respondSuccessRedirect(w http.ResponseWriter, req *http.Request, redirectGetter services.Redirector) bool {
	redirectURL := redirectGetter.Redirect()
	if redirectURL == "" {
		return false
	}

	http.Redirect(w, req, redirectURL, http.StatusTemporaryRedirect)
	return true
}

func respondSuccessRedirectPermanent(w http.ResponseWriter, req *http.Request, redirectGetter services.RedirectorPermanent) bool {
	redirectURL := redirectGetter.RedirectPermanent()
	if redirectURL == "" {
		return false
	}

	http.Redirect(w, req, redirectURL, http.StatusPermanentRedirect)
	return true
}

func respondSuccessStream(w http.ResponseWriter, streamResponse services.ContentGetter, status int) bool {
	content := streamResponse.Content()
	defer quiet.Close(content)

	headers := w.Header()
	headers.Set("Content-Type", "application/octet-stream")

	writeContentType(headers, streamResponse)
	writeContentLength(headers, streamResponse)
	writeContentRange(headers, streamResponse) // this can change Content-Length, so do this after writeContentLength()!
	writeContentFileName(headers, streamResponse)

	w.WriteHeader(status)
	_, _ = io.Copy(w, content)
	return true
}

func writeContentType(headers http.Header, streamResponse services.ContentGetter) {
	// Your stream response will just use the default content type ("application/octet-stream")
	// because you aren't capable of telling us otherwise.
	getter, ok := streamResponse.(services.ContentTypeGetter)
	if !ok {
		return
	}

	// You can tell us specifically what this byte stream contains... but you didn't.
	contentType := strings.TrimSpace(getter.ContentType())
	if contentType == "" {
		return
	}

	headers.Set("Content-Type", contentType)
}

func writeContentLength(headers http.Header, streamResponse services.ContentGetter) {
	// Only bother if the response struct can supply range information.
	getter, ok := streamResponse.(services.ContentLengthGetter)
	if !ok {
		return
	}

	// While the response can possibly supply this value, you don't appear to have done so.
	length := getter.ContentLength()
	if length <= 0 {
		return
	}

	headers.Set("Content-Length", strconv.FormatInt(int64(length), 10))
}

func writeContentRange(headers http.Header, streamResponse services.ContentGetter) {
	// Only bother if the response struct can supply range information.
	getter, ok := streamResponse.(services.ContentRangeGetter)
	if !ok {
		return
	}

	// The response can supply these values, but they don't appear to have done so for this one.
	start, end, size := getter.ContentRange()
	if start <= 0 && end <= 0 && size <= 0 {
		return
	}

	// You tried to supply meaningful values, but they're garbage.
	if end <= start || end >= size {
		return
	}

	sizeValue := "*"
	if size > 0 {
		sizeValue = strconv.FormatInt(int64(size), 10)
	}
	headers.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%s", start, end, sizeValue))
	headers.Set("Content-Length", strconv.FormatInt(int64(end-start), 10))
}

func writeContentFileName(headers http.Header, streamResponse services.ContentGetter) {
	getter, ok := streamResponse.(services.ContentFileNameGetter)
	if !ok {
		return
	}

	contentFileName := naming.CleanFileName(strings.TrimSpace(getter.ContentFileName()))
	if contentFileName == "" {
		return
	}

	headers.Set("Content-Disposition", `attachment; filename="`+contentFileName+`"`)
}

func pathParams(req *http.Request) map[string][]string {
	params := httptreemux.ContextParams(req.Context())
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	return values
}

// GatewayOption defines a setting you can apply when creating an RPC gateway via 'NewGateway'.
type GatewayOption func(*Gateway)

// WithMiddleware inserts the following chain of HTTP handlers so that they fire before
// the actual HTTP handler we generate for your service endpoint. The middleware functions
// use continuation passing style, so you should be able to plug in any off-the-shelf
// handlers like negroni.
//
// Ideally, you would NOT do any business logic in these middleware functions. It's purely
// for things like CORS which are very HTTP-specific. Anything like authorization or entity
// caching should really be done using standard services.MiddlewareFunc handlers - this way
// they fire regardless of the gateway.
func WithMiddleware(funcs ...HTTPMiddlewareFunc) GatewayOption {
	return func(gw *Gateway) {
		gw.middleware = append(gw.middleware, funcs...)
	}
}

// WithTLSConfig allows the gateway's underlying HTTP server to handle HTTPS requests using
// the configuration you provide. If you are using the Let's Encrypt auto-cert manager certificate
// configurations, this is how you can make your gateway adhere to that cert.
func WithTLSConfig(config *tls.Config) GatewayOption {
	return func(gw *Gateway) {
		gw.server.TLSConfig = config
	}
}

// WithTLSFiles allows the gateway's underlying HTTP server to handle HTTPS requests using
// the cert/key files provided.
func WithTLSFiles(certFile string, keyFile string) GatewayOption {
	return func(gw *Gateway) {
		gw.tlsCert = certFile
		gw.tlsKey = keyFile
	}
}

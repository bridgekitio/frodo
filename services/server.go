package services

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bridgekit-io/frodo/fail"
	"github.com/bridgekit-io/frodo/metadata"
)

// GatewayType is a tagging value that gateways can use to classify themselves.
type GatewayType string

// String returns the raw string value for the type.
func (t GatewayType) String() string {
	return string(t)
}

const (
	// GatewayTypeAPI marks a gateway as being an HTTP/RPC API gateway.
	GatewayTypeAPI = GatewayType("API")
	// GatewayTypeEvents marks a gateway as being event-sourced using publish/subscribe.
	GatewayTypeEvents = GatewayType("EVENTS")
)

// Gateway describes a way to execute operations on some underlying service. By
// default, service methods are closed off to all external processes, but gateways
// provide a protocol such as HTTP/RPC or PubSub to trigger them. How that actually
// happens is completely up to the Gateway implementation - the interface merely
// provides signals to start/stop the mechanism that accepts/processes function calls.
type Gateway interface {
	// Type returns the identifier used to distinguish this gateway from others registered
	// for the same service.
	Type() GatewayType
	// Register adds an ingress handler for the given operation/endpoint to this gateway.
	Register(endpoint Endpoint, route EndpointRoute)
	// Listen causes the gateway to start accepting requests to invoke methods on the
	// underlying service. Implementations should attempt to follow these rules:
	//
	//   * Block on the current Goroutine. This should unblock on an abnormal interruption
	//     to the gateway's ability to continue processing or if Shutdown() has been
	//     called elsewhere.
	//   * The error should be 'nil' if nothing has gone wrong. This is different from the
	//     behavior of http.ListenAndServe() which returns a http.ErrServerClosed error
	//     even when things shut down as expected. Gateway instances should keep their
	//     whore mouths shut and only report an error when there's actually something
	//     to be concerned about.
	Listen(ctx context.Context) error
	// Shutdown should attempt to gracefully wind down processing of requests. Where
	// possible, you should use the context to determine if/when you should give up
	// on dealing with existing work. Implementations should try to follow these rules:
	//
	//   * Immediately stop accepting incoming requests.
	//   * Allow in-process requests to finish cleanly.
	//   * Abide any cancellation of the context and give up on existing requests.
	Shutdown(ctx context.Context) error
}

// GatewayMiddleware allows gateway implementations to add special middleware to the standard
// execution pipeline for an endpoint. Your gateway would implement this when it has special
// functionality you want included in the execution of every endpoint regardless of which
// gateway is actually servicing it.
//
// The canonical example for this is the event gateway. When a service method gets invoked,
// we want to publish a "Service.Method" event afterwards no matter what so that the rest of
// your system can be notified about the event. We don't care if it was an HTTP request that
// triggered the invocation or some other event handler. We just know that we want to publish
// an event no matter what. The event middleware can return a handler(s) that injects that
// behavior by hiding it behind a generic middleware function.
type GatewayMiddleware interface {
	Gateway
	// Middleware are the functions that a Server should add to EVERY endpoint it registers.
	Middleware() MiddlewareFuncs
}

// Service encapsulates your hand-implemented service handler and includes all of the
// endpoint registration information required to power our runtime gateways.
type Service struct {
	// Name is the name of the service interface that we're building a server for.
	Name string
	// Version is an optional version identifier for this service.
	Version string
	// Handler is a reference to the actual service handler struct you provided during
	// setup of the gateway runtime in main.
	Handler any
	// Endpoints contains registration/execution information for every method/operation
	// that the service exposes.
	Endpoints []Endpoint
}

// Endpoint looks up the operation info for a method given its name. This returns the
// matching endpoint value and an 'ok' boolean similar to a map lookup.
func (svc Service) Endpoint(name string) (Endpoint, bool) {
	for _, endpoint := range svc.Endpoints {
		if endpoint.Name == name {
			return endpoint, true
		}
	}
	return Endpoint{}, false
}

// NewServer creates a new container that encapsulates one or more gateways and
// services. It helps set up endpoint routes and manages startup/shutdown routines
// so that you can start/stop accepting service requests.
//
// Example:
//
//	calcHandler := calc.CalculatorServiceHandler{}
//	calcServer := gen.NewCalculatorService(calcHandler)
//
//	server := services.NewServer(
//		services.Listen(apis.NewGateway()),
//		services.Listen(events.NewGateway()),
//		services.Register(calcServer),
//	)
func NewServer(options ...ServerOption) *Server {
	instance := Server{
		gateways:          map[GatewayType]Gateway{},
		shutdownComplete:  &sync.WaitGroup{},
		gatewayMiddleware: MiddlewareFuncs{},
		endpoints:         map[string]Endpoint{},
		onPanic: func(err error, stack []byte) {
			fmt.Printf("Panic: %v\n%v\n", err, string(stack))
		},
		logger: slog.New(slog.NewJSONHandler(nopWriter{}, nil)),
	}
	for _, option := range options {
		option(&instance)
	}

	// If any of the gateways require special processing in all of the handler (like
	// events need to publish on every invocation), capture those once.
	for _, gw := range instance.gateways {
		mw, ok := gw.(GatewayMiddleware)
		if !ok {
			continue
		}
		instance.gatewayMiddleware = append(instance.gatewayMiddleware, mw.Middleware()...)
	}

	// Now that we know our gateways are fully set up, go ahead and register the endpoint routes.
	for _, service := range instance.services {
		for _, endpoint := range service.Endpoints {
			instance.registerEndpoint(endpoint)
		}
	}
	return &instance
}

type nopWriter struct{}

func (w nopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Server is the primordial component that wrangles all of your services and gateways to get
// them talking to each other. You should not create one of these yourself. Instead, you should
// use the NewServer() constructor to do that for you.
type Server struct {
	// gateways contains the individual gateways accepting requests.
	gateways map[GatewayType]Gateway
	// services are the actual generated service endpoint handlers that we'll route requests to.
	services []*Service
	// endpoints are all of the individual operations across all registered services in this server.
	endpoints map[string]Endpoint
	// shutdownComplete waits for all gateways to be shut down before we exit.
	shutdownComplete *sync.WaitGroup
	// gatewayMiddleware aggregates all endpoint middleware functions that we want to occur on ALL
	// endpoints regardless of the gateway that's handling it.
	gatewayMiddleware MiddlewareFuncs
	// onPanic is a customizable callback that lets you perform custom logging/logic whenever the server
	// recovers from a panic that occurred during your function calls.
	onPanic OnPanicFunc
	// logger customizes how you want low-level frodo logging to be written.
	logger *slog.Logger
}

func (server *Server) registerEndpoint(endpoint Endpoint) {
	// The endpoint handler already has the user-defined middleware bound in it.
	// That should come AFTER our internal bookkeeping is complete so that their
	// handlers have everything that the framework offers at their disposal. Additionally,
	// the recovery middleware should always be the outermost handler to clean up
	// after any crap that happens anywhere else in the pipeline.
	endpoint.Handler = MiddlewareFuncs{recoverMiddleware(server.onPanic), rolesMiddleware(endpoint)}.
		Append(server.gatewayMiddleware...).
		Then(endpoint.Handler)

	server.endpoints[endpoint.QualifiedName()] = endpoint

	for _, route := range endpoint.Routes {
		if gw, ok := server.gateways[route.GatewayType]; ok {
			gw.Register(endpoint, route)
		}
	}
}

func (server *Server) Routes(gatewayType GatewayType) []EndpointRoute {
	var routes []EndpointRoute
	for _, service := range server.services {
		for _, endpoint := range service.Endpoints {
			for _, route := range endpoint.Routes {
				if route.GatewayType == gatewayType {
					routes = append(routes, route)
				}
			}
		}
	}

	slices.SortFunc(routes, func(a, b EndpointRoute) int {
		pathComp := strings.Compare(a.Path, b.Path)
		if pathComp == 0 {
			return strings.Compare(a.Method, b.Method)
		}
		return pathComp
	})
	return routes
}

// Invoke allows you to manually trigger any registered service endpoint/function given the name
// of the service/method. I'd suggest you stick to using the generated clients to invoke functions
// on your services rather than using this. This primarily exists to aid in testing - it's not really
// mean to be used in production code. But hey, you're an adult. Do what you want.
func (server *Server) Invoke(ctx context.Context, serviceName string, methodName string, req any) (any, error) {
	endpointKey := serviceName + "." + methodName
	if endpoint, ok := server.endpoints[endpointKey]; ok {
		ctx = metadata.WithRoute(ctx, metadata.EndpointRoute{
			ServiceName: serviceName,
			Name:        methodName,
			Status:      200,
		})
		return endpoint.Handler(ctx, req)
	}
	return nil, fail.NotFound("server operation not found: %s", endpointKey)
}

// Run turns on every gateway currently assigned to this service runtime. Call this
// once your service setup and registration is complete in order to start accepting
// incoming requests through your gateway(s).
func (server *Server) Run(ctx context.Context) error {
	server.shutdownComplete.Add(1)

	errs, _ := fail.NewGroup(ctx)
	for _, gw := range server.gateways {
		server.logger.Info("[frodo] starting gateway: " + gw.Type().String())
		errs.Go(func() error { return gw.Listen(ctx) })
	}

	// We had an issue starting up the server, so just get out and
	// let the user determine how to handle the fact that the HTTP
	// server or event broker didn't work.
	if err := errs.Wait(); err != nil {
		server.shutdownComplete.Done()
		return err
	}

	// You called Shutdown() is going as expected, so nons of our listener
	// returned an error as they just stopped listening for new work. We, however,
	// need to wait for any already-in-process requests to finish up. This should
	// unblock when the work is done or when we hit our graceful deadline and give up.
	server.shutdownComplete.Wait()
	return nil
}

// Shutdown attempts to gracefully shut down all of the gateways associated with this
// service runtime. It should immediately stop accepting new requests and then wait
// for existing requests to finish before returning. The context should be used to
// provide a cancellation/timeout to limit how long this will wait for in-flight
// requests to finish up.
func (server *Server) Shutdown(ctx context.Context) error {
	defer server.shutdownComplete.Done()

	errs, _ := fail.NewGroup(ctx)
	for _, gw := range server.gateways {
		// Bug avoidance note: I'm capturing the shutdown method outside
		// the goroutine function because if we just called gw.Shutdown() in
		// the anonymous function, every iteration would just call shutdown
		// on the first gateway over and over. That's a side effect of the
		// semantics of the variables in Go's for loops.
		shutdown := gw.Shutdown
		errs.Go(func() error { return shutdown(ctx) })
	}
	return errs.Wait()
}

// ShutdownOnInterrupt provides some convenience around shutting down this service.
// This function will block until the process either receives a SIGTERM or SIGINT
// signal. At that point, it will invoke Shutdown() whose context will have a
// deadline of the given duration.
//
// Example:
//
//	// Ignore the bad (non-existent) error handling, but here's how your server
//	// setup/teardown code looks using ShutdownOnInterrupt. The server will start
//	// up and run until the process gets a SIGINT/SIGTERM signal. At that point, it
//	// will give all in-process requests in all gateways 10 seconds to finish.
//	void main() {
//		server := services.NewServer(...)
//		go server.ShutdownOnInterrupt(10*time.Second)
//		server.Run(context.Background())
//	}
func (server *Server) ShutdownOnInterrupt(gracefulTimeout time.Duration) {
	// Block while waiting for a SIGTERM or SIGINT signal from the shell.
	var interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-interrupt

	// This context ensures that we give the gateways some time to finish
	// up their in-process requests before shutting down.
	ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("error shutting down services: %v", err)
	}
}

// ServerOption defines a setting that you can change on a services.Server while
// setting up your application in main().
type ServerOption func(*Server)

// Listen adds another gateway to the server. You can supply this option more
// than once in order to provide multiple types of gateways. For instance, you can
// call it once to provide settings for an API/HTTP gateway and again to provide
// settings for an event source gateway.
func Listen(gw Gateway) ServerOption {
	return func(server *Server) {
		server.gateways[gw.Type()] = gw
	}
}

// Register adds endpoint handlers for the given service(s) to the appropriate gateways.
// Typically, you don't create the Service pointer yourself. These are built for you when
// you use the code generation tools to build the gateways based on your service interfaces.
//
// Consider this call akin to adding routes to an HTTP router. This just adds routes to every
// applicable gateway in your runtime server.
func Register(services ...*Service) ServerOption {
	return func(server *Server) {
		server.services = append(server.services, services...)
	}
}

// OnPanicFunc is the signature for custom callbacks to invoke when a panic occurs in your service code.
type OnPanicFunc func(err error, stack []byte)

// OnPanic provides a custom callback that will let you log/observe the error/stack from any panic that occurred
// in your service method code.
func OnPanic(handler OnPanicFunc) ServerOption {
	return func(server *Server) {
		server.onPanic = handler
	}
}

// WithLogger customizes the logger used by the server to output various bits of debugging info.
func WithLogger(logger *slog.Logger) ServerOption {
	return func(server *Server) {
		server.logger = logger
	}
}

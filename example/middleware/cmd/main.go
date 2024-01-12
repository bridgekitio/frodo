package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bridgekitio/frodo/example/basic/calc"
	gen "github.com/bridgekitio/frodo/example/basic/calc/gen"
	"github.com/bridgekitio/frodo/fail"
	"github.com/bridgekitio/frodo/metadata"
	"github.com/bridgekitio/frodo/services"
	"github.com/bridgekitio/frodo/services/gateways/apis"
	"github.com/bridgekitio/frodo/services/gateways/events"
)

func main() {
	fmt.Println("Initializing server")

	// Set up your service the same as before, just with a varargs list of middleware functions to execute. Every
	// method in the service will invoke the logger and the Monday check as part of their request flows.
	//
	// These middleware functions fire whether a method was called using the HTTP API or asynchronously using
	// the event gateway. It doesn't matter.
	//
	// Notice that you're setting these middleware functions on the server for the specific service. This means
	// that if you're running multiple services together, each one can have different middleware chains if you like.
	calcHandler := calc.CalculatorServiceHandler{}
	calcService := gen.CalculatorServiceServer(calcHandler, LoggerMiddleware, NoMathMondayMiddleware)

	// If you need HTTP-specific middleware for things like CORS or content length limiting, you specify
	// those when setting up the API gateway.
	server := services.NewServer(
		services.Listen(apis.NewGateway(":8080", apis.WithMiddleware(
			FilterUserAgent,
			CORS,
		))),
		services.Listen(events.NewGateway()),
		services.Register(calcService),
	)

	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("Quick examples:")
	fmt.Println("  curl http://localhost:8080/add/5/12")
	fmt.Println("  curl http://localhost:8080/sub/15/9")

	// Fire up the API and shut down gracefully when we receive a SIGINT or SIGTERM signal.
	go server.ShutdownOnInterrupt(10 * time.Second)
	if err := server.Run(); err != nil {
		panic(err)
	}

	fmt.Println("Bye bye...")
}

/*
 * ===== Service Level Middleware =====
 */

// LoggerMiddleware outputs the start/end of every service invocation that occurs. This shows you how you can
// inject behavior before and after your method invocation. Here, it's used for logging, but you can do this for
// metrics collection resource opening/closing, etc.
func LoggerMiddleware(ctx context.Context, req any, next services.HandlerFunc) (any, error) {
	route := metadata.Route(ctx)
	fmt.Printf("[%s] Invoking\n", route.QualifiedName())
	res, err := next(ctx, req)
	fmt.Printf("[%s] Complete (success=%v)\n", route.QualifiedName(), err == nil)
	return res, err
}

// NoMathMondayMiddleware interrupts the normal service method invocation if you attempt to call it
// when today is a Monday. Garfield approves. Ultimately, this shows you how you can create general
// purpose middleware that can short circuit your service invocation. This is useful for cases like
// authentication/authorization where you don't want to let the invocation occur unless some
// criteria is met.
func NoMathMondayMiddleware(ctx context.Context, req any, next services.HandlerFunc) (any, error) {
	if time.Now().Weekday() == time.Monday {
		return nil, fail.PermissionDenied("We don't do math on Mondays")
	}
	return next(ctx, req)
}

/*
 * ===== HTTP-Specific Middleware =====
 */

func CORS(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	switch req.Method {
	case http.MethodOptions:
		// Do crazy CORS preflight logic here...
	default:
		next(w, req) // Just a normal request. Process like normal.
	}
}

func FilterUserAgent(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// Totally fool-proof way to stop bots from accessing the service!!!
	if strings.Contains(req.UserAgent(), "Bot") {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	next(w, req) // invoke the route/method like normal.
}

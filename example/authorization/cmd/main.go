package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bridgekitio/frodo/example/authorization/sensitive"
	gen "github.com/bridgekitio/frodo/example/authorization/sensitive/gen"
	"github.com/bridgekitio/frodo/fail"
	"github.com/bridgekitio/frodo/internal/slices"
	"github.com/bridgekitio/frodo/metadata"
	"github.com/bridgekitio/frodo/services"
	"github.com/bridgekitio/frodo/services/gateways/apis"
)

func main() {
	fmt.Println("Initializing server")

	// Apply authentication/authorization middleware to guard every endpoint with checks that ensure
	// the caller has valid credentials and that the user is actually allowed to access that resource.
	service := gen.SecretServiceServer(&sensitive.SecretServiceHandler{}, Authenticate, Authorize)
	server := services.NewServer(
		services.Listen(apis.NewGateway(":8080")),
		services.Register(service),
	)

	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("Quick examples:")
	fmt.Println("  curl http://localhost:8080/group/123")
	fmt.Println("  curl -H 'Authorization: Bearer admin' http://localhost:8080/group/123")
	fmt.Println("  curl -H 'Authorization: Bearer 123'   http://localhost:8080/group/123")
	fmt.Println("  curl -H 'Authorization: Bearer 456'   http://localhost:8080/group/123")
	fmt.Println("  curl -H 'Authorization: Bearer XXX'   http://localhost:8080/group/123")

	// Fire up the API and shut down gracefully when we receive a SIGINT or SIGTERM signal.
	go server.ShutdownOnInterrupt(10 * time.Second)
	if err := server.Run(context.Background()); err != nil {
		panic(err)
	}
	fmt.Println("Bye bye...")
}

func Authenticate(ctx context.Context, req any, next services.HandlerFunc) (any, error) {
	// Don't judge. The important thing to note is that the call to `metadata.Authorization()` grabs the
	// underlying HTTP Authorization header value, and you can decipher that however you like. What you
	// should take away is the notion that we can take the callers credentials and map them to
	// some sort of calling context that can be validated later.
	switch metadata.Authorization(ctx) {
	case "Bearer admin":
		ctx = contextWithUser(ctx, &User{Roles: []string{"admin.read", "admin.write"}})
		return next(ctx, req)
	case "Bearer 123":
		ctx = contextWithUser(ctx, &User{Roles: []string{"group.123.read", "group.123.write"}})
		return next(ctx, req)
	case "Bearer 456":
		ctx = contextWithUser(ctx, &User{Roles: []string{"group.456.read", "group.456.write"}})
		return next(ctx, req)
	default:
		return nil, fail.BadCredentials("invalid token")
	}
}

func Authorize(ctx context.Context, req any, next services.HandlerFunc) (any, error) {
	// If you hit the endpoint '/group/123' then this will have the roles "group.123.read"
	// and "group.123.write". If you hit the endpoint '/group/789' then this will have "group.789.read"
	// as well as "group.789.write". These are the fully resolved values that came from the ROLES doc
	// option in secret_service.go.
	endpointRoles := metadata.Route(ctx).Roles

	// This should have been populated by the Authenticate middleware.
	caller := userFromContext(ctx)

	// Yes, you have the rights to access this endpoint.
	if caller.HasAnyRole(endpointRoles) {
		return next(ctx, req)
	}
	return nil, fail.PermissionDenied("you don't have rights to access this resource, bub")
}

/*
 * Keep track of the authenticated/authorized user in the context, so you can decouple that from
 * your actual service logic. It also helps you isolate your identity management logic better.
 */

type userContextKey struct{}

func userFromContext(ctx context.Context) *User {
	return ctx.Value(userContextKey{}).(*User)
}

func contextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

type User struct {
	Roles []string
}

func (user *User) HasAnyRole(roles []string) bool {
	for _, role := range roles {
		if slices.Contains(user.Roles, role) {
			return true
		}
	}
	return false
}

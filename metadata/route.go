package metadata

import (
	"context"
)

type contextKeyRoute struct{}

// Route extracts the info about the current operation/endpoint being invoked.
func Route(ctx context.Context) EndpointRoute {
	if ctx == nil {
		return EndpointRoute{}
	}
	if route, ok := ctx.Value(contextKeyRoute{}).(EndpointRoute); ok {
		return route
	}
	return EndpointRoute{}
}

// WithRoute stores the current operation/endpoint being invoked. Typically,
// you will not need to call this yourself as the framework will take care of this
// for you. You just need to call metadata.Endpoint() when you need to check that value.
func WithRoute(ctx context.Context, route EndpointRoute) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, contextKeyRoute{}, route)
}

// EndpointRoute stores a subset of the master endpoint information that you might
// want for logging and lookup. It does not include any of the actual functionality
// fields such as handlers/factories.
type EndpointRoute struct {
	// ServiceName is the name of the service that this operation is part of.
	ServiceName string
	// Name is the name of the function/operation that this endpoint describes.
	Name string
	// Type indicates the type of gateway/route that triggered the operation to be
	// invoked in the first place. This is a way for you to check if the service is
	// being handled due to an API call or some sort of event handler.
	Type string
	// Method describes some sort of action/verb that describes this route. For API endpoints
	// it is the HTTP method (e.g. GET, PUT, POST, etc). For events, it is "ON", and so forth.
	Method string
	// Path describes the actual unique routing path that the gateway should use to ensure
	// that requests get to this endpoint. For API endpoints, it's the request path
	// like "/user/{ID}" and for event endpoints, it's the subscription key like "FooService.Save".
	Path string
	// Status passes along the route's HTTP ### value from doc options when applicable. This is
	// automatically set to 200 for event-based routes for a consistent "OK nothing went wrong" default.
	Status int
	// Roles are used for role-based security where you can say that this endpoint requires the user/caller
	// to have either "admin.write" or "group.write" privileges. When defining your services, you can parameterize
	// your roles such as "group.{ID}.write", but the ones stored in this slice should have already been
	// resolved with the appropriate runtime values (e.g. "group.123.write").
	//
	// Friendly reminder that these are the roles you want the security layer to look for - it's
	// not necessarily what the caller actually has!
	Roles []string
}

// QualifiedName returns the fully-qualified name/identifier of this service operation. It
// is simply the formatted string "ServiceName.MethodName".
func (e EndpointRoute) QualifiedName() string {
	switch {
	case e.ServiceName == "":
		return e.Name
	case e.Name == "":
		return e.ServiceName
	default:
		return e.ServiceName + "." + e.Name
	}
}

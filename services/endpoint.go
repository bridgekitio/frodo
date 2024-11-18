package services

import (
	"context"
)

// HandlerFunc is the general purpose signature for any endpoint handler.
type HandlerFunc func(ctx context.Context, req any) (any, error)

// StructPointer is a tagging type used to indicate either a pointer to the "request" struct
// to service methods or a pointer to the "response" struct of the method.
type StructPointer any

// Endpoint describes an operation on an underlying service that we expose through one of
// potentially multiple gateways. The Routes on this Endpoint indicate all of the ingress
// mechanisms available to invoke this operation.
type Endpoint struct {
	// ServiceName is the name of the service that this operation is part of.
	ServiceName string
	// Name is the name of the function/operation that this endpoint describes.
	Name string
	// Handler is the actual function that will trigger the work on then underlying
	// service to accomplish the work that this endpoint is intended to.
	Handler HandlerFunc
	// NewInput returns a pointer to the struct that we pass into the handler function
	// for this input (e.g. "*LoginRequest"). It is the request struct pointer.
	NewInput func() StructPointer
	// Roles helps support role-based security by defining role patterns to indicate which
	// users are allowed to access this endpoint.  For example:
	//
	//    endpoint.Roles = []string{
	//        "admin.write",
	//        "group.{Group.ID}.write",
	//    }
	//
	// Notice that the roles should be allowed to have path variables that we can fill in
	// at runtime with the incoming binding data.
	Roles []string
	// Routes defines the actual ingress routes that allow this service operation to
	// be invoked by various gateways. For instance, they tell you that you can invoke
	// the API call "GET /user/{ID}" to invoke it or that it should trigger when the
	// event "ON UserService.UserCreated" fires.
	Routes []EndpointRoute
}

// QualifiedName returns the fully-qualified name/identifier of this service operation. It
// is simply the formatted string "ServiceName.MethodName".
func (end Endpoint) QualifiedName() string {
	return end.ServiceName + "." + end.Name
}

// EndpointRoute defines an actual ingress route that allows a service operation to
// be invoked by various gateways. For instance, one route will tell you that you can invoke
// the API call "GET /user/{ID}" to invoke the method or that it should run when the
// event "ON UserService.UserCreated" fires.
type EndpointRoute struct {
	// GatewayType indicates the type of gateway that should service this route. For instance
	// if the type is "HTTP" then the API gateway should take care of it. If the type
	// is "EVENT" then we should let our pub-sub event source take care of it.
	GatewayType GatewayType
	// Method describes some sort of action/verb that describes this route. For API endpoints
	// it is the HTTP method (e.g. GET, PUT, POST, etc). For events it is "ON", and so forth.
	Method string
	// Path describes the actual unique routing path that the gateway should use to ensure
	// that requests get to this endpoint. For API endpoints, it's the request path
	// like "/user/{ID}" and for event endpoints, it's the subscription key like "FooService.Save".
	Path string
	// PathParams contains the names of the path variables/parameters you expect in the path of
	// this endpoint. For instance, the path "/user/{UserID}/transaction/{TransactionID}" would set
	// this slice to []string{"UserID", "TransactionID"}. This allows you to quickly bind only the
	// values you expect in the pattern.
	PathParams []string
	// Group provides additional routing/grouping info that means different things to different gateways.
	Group string
	// Status is mainly used by API gateway routes to determine what HTTP status code we should
	// return to the caller when this endpoint succeeds. By default, this is 200.
	Status int
	// Roles helps support role-based security by defining role patterns to indicate which
	// users are allowed to access this endpoint. This is the same as the Roles in the parent Endpoint that
	// this route belongs to.
	Roles []string
	// ServiceName is the name of the service that this operation is part of.
	ServiceName string
	// Name is the name of the function/operation that this endpoint describes.
	Name string
}

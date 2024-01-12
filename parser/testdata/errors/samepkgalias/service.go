package samepkgalias

import (
	"context"
)

/*
 * The service is fine, but I don't have a solution built for defining request/response structs that
 * are just type references to other things in this package. Oddly, they work when you alias something
 * from another package, but not when the aliased type is in the same package as the service definition.
 *
 * Parsing Response should fail.
 */

type FooService interface {
	Hello(context.Context, *Request) (*Response, error)
}

type Request struct {
	ID string
}

type Response Model

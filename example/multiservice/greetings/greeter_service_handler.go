package greetings

import (
	"context"

	"github.com/bridgekit-io/frodo/fail"
)

type GreeterServiceHandler struct{}

func (GreeterServiceHandler) Greet(_ context.Context, req *GreetRequest) (*GreetResponse, error) {
	if req.Name == "" {
		return nil, fail.BadRequest("name is required")
	}
	return &GreetResponse{Value: "Hello, " + req.Name}, nil
}

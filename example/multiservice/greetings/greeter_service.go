package greetings

import "context"

type GreeterService interface {
	Greet(ctx context.Context, req *GreetRequest) (*GreetResponse, error)
}

type GreetRequest struct {
	Name string
}

type GreetResponse struct {
	Value string
}

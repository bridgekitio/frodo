package gen

import (
	"context"

	"github.com/bridgekitio/frodo/example/multiservice/greetings"
	"github.com/bridgekitio/frodo/fail"
	"github.com/bridgekitio/frodo/services"
)

func NewGreeterService(handler greetings.GreeterService, middleware ...services.MiddlewareFunc) *services.Service {
	middlewareFuncs := services.MiddlewareFuncs(middleware)

	return &services.Service{
		Name:    "GreeterService",
		Version: "0.0.1",
		Handler: handler,
		Endpoints: []services.Endpoint{
			{
				ServiceName: "GreeterService",
				Name:        "Greet",
				NewInput:    func() services.StructPointer { return &greetings.GreetRequest{} },
				Handler: middlewareFuncs.Then(func(ctx context.Context, req any) (any, error) {
					typedReq, ok := req.(*greetings.GreetRequest)
					if !ok {
						return nil, fail.Unexpected("invalid request argument type")
					}
					return handler.Greet(ctx, typedReq)
				}),
				Routes: []services.EndpointRoute{
					{
						GatewayType: services.GatewayTypeAPI,
						Method:      "POST",
						Path:        "GreeterService.Greet",
						Status:      200,
					},
				},
			},
		},
	}
}

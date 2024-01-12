package gen

import (
	"context"

	"github.com/bridgekitio/frodo/example/multiservice/dismissal"
	"github.com/bridgekitio/frodo/fail"
	"github.com/bridgekitio/frodo/services"
)

func NewDismissService(handler dismissal.DismissService, middleware ...services.MiddlewareFunc) *services.Service {
	middlewareFuncs := services.MiddlewareFuncs(middleware)

	return &services.Service{
		Name:    "DismissService",
		Version: "0.0.1",
		Handler: handler,
		Endpoints: []services.Endpoint{
			{
				ServiceName: "DismissService",
				Name:        "Dismiss",
				NewInput:    func() services.StructPointer { return &dismissal.DismissRequest{} },
				Handler: middlewareFuncs.Then(func(ctx context.Context, req any) (any, error) {
					typedReq, ok := req.(*dismissal.DismissRequest)
					if !ok {
						return nil, fail.Unexpected("invalid request argument type")
					}
					return handler.Dismiss(ctx, typedReq)
				}),
				Routes: []services.EndpointRoute{
					{
						GatewayType: services.GatewayTypeAPI,
						Method:      "POST",
						Path:        "DismissService.Dismiss",
						Status:      200,
					},
				},
			},
		},
	}
}

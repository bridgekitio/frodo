// Code generated by Frodo - DO NOT EDIT.
//
//	Timestamp: Tue, 05 Mar 2024 08:16:00 EST
//	Source:    other_service.go
//	Generator: https://github.com/bridgekitio/frodo
package testext

import (
	"context"

	"github.com/bridgekitio/frodo/fail"
	"github.com/bridgekitio/frodo/internal/testext"
	"github.com/bridgekitio/frodo/services"
)

// OtherServiceServer accepts your "real" OtherService instance (the thing that really does
// the work), and returns a set of endpoint routes which allow this service to be consumed
// via the gateways/listeners you configure in main().
//
//	// Example
//	serviceHandler := testext.OtherServiceHandler{ /* set up to your liking */ }
//	server := services.New(
//		services.Listen(apis.NewGateway()),
//		services.Register(testextgen.OtherServiceServer(serviceHandler)),
//	)
//	server.Listen()
//
// From there, you can add middleware, event sourcing support and more. Look at the frodo
// documentation for more details/examples on how to make your service production ready.
func OtherServiceServer(handler testext.OtherService, middleware ...services.MiddlewareFunc) *services.Service {
	middlewareFuncs := services.MiddlewareFuncs(middleware)

	return &services.Service{
		Name:    "OtherService",
		Version: "0.0.1",
		Handler: handler,
		Endpoints: []services.Endpoint{

			{
				ServiceName: "OtherService",
				Name:        "ChainFail",
				NewInput:    func() services.StructPointer { return &testext.OtherRequest{} },
				Handler: middlewareFuncs.Then(func(ctx context.Context, req any) (any, error) {
					typedReq, ok := req.(*testext.OtherRequest)
					if !ok {
						return nil, fail.Unexpected("invalid request argument type")
					}
					return handler.ChainFail(ctx, typedReq)
				}),
				Roles: []string{},
				Routes: []services.EndpointRoute{
					{
						GatewayType: "API",
						Method:      "POST",
						Path:        "/OtherService.ChainFail",
						PathParams:  []string{},
						Status:      200,
					},

					{
						GatewayType: "EVENTS",
						Method:      "ON",
						Path:        "OtherService.ChainOne",
						PathParams:  []string{},
						Status:      0,
					},
				},
			},

			{
				ServiceName: "OtherService",
				Name:        "ChainFailAfter",
				NewInput:    func() services.StructPointer { return &testext.OtherRequest{} },
				Handler: middlewareFuncs.Then(func(ctx context.Context, req any) (any, error) {
					typedReq, ok := req.(*testext.OtherRequest)
					if !ok {
						return nil, fail.Unexpected("invalid request argument type")
					}
					return handler.ChainFailAfter(ctx, typedReq)
				}),
				Roles: []string{},
				Routes: []services.EndpointRoute{
					{
						GatewayType: "API",
						Method:      "POST",
						Path:        "/OtherService.ChainFailAfter",
						PathParams:  []string{},
						Status:      200,
					},

					{
						GatewayType: "EVENTS",
						Method:      "ON",
						Path:        "OtherService.ChainFail",
						PathParams:  []string{},
						Status:      0,
					},
				},
			},

			{
				ServiceName: "OtherService",
				Name:        "ChainFour",
				NewInput:    func() services.StructPointer { return &testext.OtherRequest{} },
				Handler: middlewareFuncs.Then(func(ctx context.Context, req any) (any, error) {
					typedReq, ok := req.(*testext.OtherRequest)
					if !ok {
						return nil, fail.Unexpected("invalid request argument type")
					}
					return handler.ChainFour(ctx, typedReq)
				}),
				Roles: []string{},
				Routes: []services.EndpointRoute{
					{
						GatewayType: "API",
						Method:      "POST",
						Path:        "/OtherService.ChainFour",
						PathParams:  []string{},
						Status:      200,
					},

					{
						GatewayType: "EVENTS",
						Method:      "ON",
						Path:        "OtherService.ChainTwo",
						PathParams:  []string{},
						Status:      0,
					},
				},
			},

			{
				ServiceName: "OtherService",
				Name:        "ChainOne",
				NewInput:    func() services.StructPointer { return &testext.OtherRequest{} },
				Handler: middlewareFuncs.Then(func(ctx context.Context, req any) (any, error) {
					typedReq, ok := req.(*testext.OtherRequest)
					if !ok {
						return nil, fail.Unexpected("invalid request argument type")
					}
					return handler.ChainOne(ctx, typedReq)
				}),
				Roles: []string{},
				Routes: []services.EndpointRoute{
					{
						GatewayType: "API",
						Method:      "POST",
						Path:        "/OtherService.ChainOne",
						PathParams:  []string{},
						Status:      200,
					},
				},
			},

			{
				ServiceName: "OtherService",
				Name:        "ChainThree",
				NewInput:    func() services.StructPointer { return &testext.OtherRequest{} },
				Handler: middlewareFuncs.Then(func(ctx context.Context, req any) (any, error) {
					typedReq, ok := req.(*testext.OtherRequest)
					if !ok {
						return nil, fail.Unexpected("invalid request argument type")
					}
					return handler.ChainThree(ctx, typedReq)
				}),
				Roles: []string{},
				Routes: []services.EndpointRoute{
					{
						GatewayType: "API",
						Method:      "POST",
						Path:        "/OtherService.ChainThree",
						PathParams:  []string{},
						Status:      200,
					},

					{
						GatewayType: "EVENTS",
						Method:      "ON",
						Path:        "OtherService.ChainTwo",
						PathParams:  []string{},
						Status:      0,
					},
				},
			},

			{
				ServiceName: "OtherService",
				Name:        "ChainTwo",
				NewInput:    func() services.StructPointer { return &testext.OtherRequest{} },
				Handler: middlewareFuncs.Then(func(ctx context.Context, req any) (any, error) {
					typedReq, ok := req.(*testext.OtherRequest)
					if !ok {
						return nil, fail.Unexpected("invalid request argument type")
					}
					return handler.ChainTwo(ctx, typedReq)
				}),
				Roles: []string{},
				Routes: []services.EndpointRoute{
					{
						GatewayType: "API",
						Method:      "POST",
						Path:        "/OtherService.ChainTwo",
						PathParams:  []string{},
						Status:      200,
					},

					{
						GatewayType: "EVENTS",
						Method:      "ON",
						Path:        "OtherService.ChainOne",
						PathParams:  []string{},
						Status:      0,
					},
				},
			},

			{
				ServiceName: "OtherService",
				Name:        "ListenWell",
				NewInput:    func() services.StructPointer { return &testext.OtherRequest{} },
				Handler: middlewareFuncs.Then(func(ctx context.Context, req any) (any, error) {
					typedReq, ok := req.(*testext.OtherRequest)
					if !ok {
						return nil, fail.Unexpected("invalid request argument type")
					}
					return handler.ListenWell(ctx, typedReq)
				}),
				Roles: []string{},
				Routes: []services.EndpointRoute{
					{
						GatewayType: "API",
						Method:      "POST",
						Path:        "/OtherService.ListenWell",
						PathParams:  []string{},
						Status:      200,
					},

					{
						GatewayType: "EVENTS",
						Method:      "ON",
						Path:        "OtherService.SpaceOut",
						PathParams:  []string{},
						Status:      0,
					},

					{
						GatewayType: "EVENTS",
						Method:      "ON",
						Path:        "SampleService.TriggerUpperCase",
						PathParams:  []string{},
						Status:      0,
					},
				},
			},

			{
				ServiceName: "OtherService",
				Name:        "RPCExample",
				NewInput:    func() services.StructPointer { return &testext.OtherRequest{} },
				Handler: middlewareFuncs.Then(func(ctx context.Context, req any) (any, error) {
					typedReq, ok := req.(*testext.OtherRequest)
					if !ok {
						return nil, fail.Unexpected("invalid request argument type")
					}
					return handler.RPCExample(ctx, typedReq)
				}),
				Roles: []string{},
				Routes: []services.EndpointRoute{
					{
						GatewayType: "API",
						Method:      "POST",
						Path:        "/OtherService.RPCExample",
						PathParams:  []string{},
						Status:      200,
					},
				},
			},

			{
				ServiceName: "OtherService",
				Name:        "SpaceOut",
				NewInput:    func() services.StructPointer { return &testext.OtherRequest{} },
				Handler: middlewareFuncs.Then(func(ctx context.Context, req any) (any, error) {
					typedReq, ok := req.(*testext.OtherRequest)
					if !ok {
						return nil, fail.Unexpected("invalid request argument type")
					}
					return handler.SpaceOut(ctx, typedReq)
				}),
				Roles: []string{},
				Routes: []services.EndpointRoute{
					{
						GatewayType: "API",
						Method:      "POST",
						Path:        "/OtherService.SpaceOut",
						PathParams:  []string{},
						Status:      200,
					},
				},
			},
		},
	}
}

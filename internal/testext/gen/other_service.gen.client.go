// Code generated by Frodo - DO NOT EDIT.
//
//	Timestamp: Wed, 20 Mar 2024 10:42:32 EDT
//	Source:    other_service.go
//	Generator: https://github.com/bridgekitio/frodo
package testext

import (
	"context"

	"github.com/bridgekitio/frodo/fail"
	"github.com/bridgekitio/frodo/internal/testext"
	"github.com/bridgekitio/frodo/services/clients"
)

// OtherServiceClient creates an RPC client that conforms to the OtherService interface, but delegates
// work to remote instances. You must supply the base address of the remote service gateway instance or
// the load balancer for that service.
//
// OtherService primarily exists to show that we can send event signals between services.
func OtherServiceClient(address string, options ...clients.ClientOption) testext.OtherService {
	serviceClient := clients.NewClient("OtherService", address, options...)
	return &otherServiceClient{Client: serviceClient}
}

// otherServiceClient manages all interaction w/ a remote OtherService instance by letting you invoke functions
// on this instance as if you were doing it locally (hence... RPC client). Use the OtherServiceClient constructor
// function to actually get an instance of this client.
type otherServiceClient struct {
	clients.Client
}

// ChainFail fires after ChainOne, but should always return an error. This will prevent ChainFailAfter
// from ever actually running.
func (client *otherServiceClient) ChainFail(ctx context.Context, request *testext.OtherRequest) (*testext.OtherResponse, error) {

	if ctx == nil {
		return nil, fail.Unexpected("precondition failed: nil context")
	}
	if request == nil {
		return nil, fail.Unexpected("precondition failed: nil request")
	}

	response := &testext.OtherResponse{}
	err := client.Invoke(ctx, "POST", "/OtherService.ChainFail", request, response)
	return response, err

}

// ChainFailAfter is dependent on a successful call to ChainFail... which always fails. So this NEVER runs.
func (client *otherServiceClient) ChainFailAfter(ctx context.Context, request *testext.OtherRequest) (*testext.OtherResponse, error) {

	if ctx == nil {
		return nil, fail.Unexpected("precondition failed: nil context")
	}
	if request == nil {
		return nil, fail.Unexpected("precondition failed: nil request")
	}

	response := &testext.OtherResponse{}
	err := client.Invoke(ctx, "POST", "/OtherService.ChainFailAfter", request, response)
	return response, err

}

// ChainFour is used to test that methods invoked via the event gateway can trigger even more events.
func (client *otherServiceClient) ChainFour(ctx context.Context, request *testext.OtherRequest) (*testext.OtherResponse, error) {

	if ctx == nil {
		return nil, fail.Unexpected("precondition failed: nil context")
	}
	if request == nil {
		return nil, fail.Unexpected("precondition failed: nil request")
	}

	response := &testext.OtherResponse{}
	err := client.Invoke(ctx, "POST", "/OtherService.ChainFour", request, response)
	return response, err

}

// ChainOne allows us to test the cascading of events to create more complex flows. When this
// finishes it will trigger ChainTwo which will, in turn, trigger ChainThree and ChainFour.
func (client *otherServiceClient) ChainOne(ctx context.Context, request *testext.OtherRequest) (*testext.OtherResponse, error) {

	if ctx == nil {
		return nil, fail.Unexpected("precondition failed: nil context")
	}
	if request == nil {
		return nil, fail.Unexpected("precondition failed: nil request")
	}

	response := &testext.OtherResponse{}
	err := client.Invoke(ctx, "POST", "/OtherService.ChainOne", request, response)
	return response, err

}

// ChainThree is used to test that methods invoked via the event gateway can trigger even more events.
func (client *otherServiceClient) ChainThree(ctx context.Context, request *testext.OtherRequest) (*testext.OtherResponse, error) {

	if ctx == nil {
		return nil, fail.Unexpected("precondition failed: nil context")
	}
	if request == nil {
		return nil, fail.Unexpected("precondition failed: nil request")
	}

	response := &testext.OtherResponse{}
	err := client.Invoke(ctx, "POST", "/OtherService.ChainThree", request, response)
	return response, err

}

// ChainTwo is used to test that methods invoked via the event gateway can trigger even more events.
func (client *otherServiceClient) ChainTwo(ctx context.Context, request *testext.OtherRequest) (*testext.OtherResponse, error) {

	if ctx == nil {
		return nil, fail.Unexpected("precondition failed: nil context")
	}
	if request == nil {
		return nil, fail.Unexpected("precondition failed: nil request")
	}

	response := &testext.OtherResponse{}
	err := client.Invoke(ctx, "POST", "/OtherService.ChainTwo", request, response)
	return response, err

}

// ListenWell can listen for successful responses across multiple services.
func (client *otherServiceClient) ListenWell(ctx context.Context, request *testext.OtherRequest) (*testext.OtherResponse, error) {

	if ctx == nil {
		return nil, fail.Unexpected("precondition failed: nil context")
	}
	if request == nil {
		return nil, fail.Unexpected("precondition failed: nil request")
	}

	response := &testext.OtherResponse{}
	err := client.Invoke(ctx, "POST", "/OtherService.ListenWell", request, response)
	return response, err

}

// RPCExample invokes the TriggerUpperCase() function on the SampleService to get work done.
// This will make sure that we can do cross-service communication.
func (client *otherServiceClient) RPCExample(ctx context.Context, request *testext.OtherRequest) (*testext.OtherResponse, error) {

	if ctx == nil {
		return nil, fail.Unexpected("precondition failed: nil context")
	}
	if request == nil {
		return nil, fail.Unexpected("precondition failed: nil request")
	}

	response := &testext.OtherResponse{}
	err := client.Invoke(ctx, "POST", "/OtherService.RPCExample", request, response)
	return response, err

}

// SpaceOut takes your input text and puts spaces in between all the letters.
func (client *otherServiceClient) SpaceOut(ctx context.Context, request *testext.OtherRequest) (*testext.OtherResponse, error) {

	if ctx == nil {
		return nil, fail.Unexpected("precondition failed: nil context")
	}
	if request == nil {
		return nil, fail.Unexpected("precondition failed: nil request")
	}

	response := &testext.OtherResponse{}
	err := client.Invoke(ctx, "POST", "/OtherService.SpaceOut", request, response)
	return response, err

}

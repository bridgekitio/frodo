package testext

import "context"

//go:generate ../../out/frodo server  $GOFILE
//go:generate ../../out/frodo client  $GOFILE
//go:generate ../../out/frodo client  $GOFILE --language=js
//go:generate ../../out/frodo client  $GOFILE --language=dart

// OtherService primarily exists to show that we can send event signals between services.
type OtherService interface {
	// SpaceOut takes your input text and puts spaces in between all the letters.
	SpaceOut(context.Context, *OtherRequest) (*OtherResponse, error)

	// RPCExample invokes the TriggerUpperCase() function on the SampleService to get work done.
	// This will make sure that we can do cross-service communication.
	RPCExample(context.Context, *OtherRequest) (*OtherResponse, error)

	// ListenWell can listen for successful responses across multiple services.
	//
	// ON OtherService.SpaceOut
	// ON SampleService.TriggerUpperCase
	ListenWell(context.Context, *OtherRequest) (*OtherResponse, error)

	// ChainOne allows us to test the cascading of events to create more complex flows. When this
	// finishes it will trigger ChainTwo which will, in turn, trigger ChainThree and ChainFour.
	ChainOne(ctx context.Context, request *OtherRequest) (*OtherResponse, error)

	// ChainTwo is used to test that methods invoked via the event gateway can trigger even more events.
	//
	// ON OtherService.ChainOne
	ChainTwo(ctx context.Context, request *OtherRequest) (*OtherResponse, error)

	// ChainThree is used to test that methods invoked via the event gateway can trigger even more events.
	//
	// ON OtherService.ChainTwo
	ChainThree(ctx context.Context, request *OtherRequest) (*OtherResponse, error)

	// ChainFour is used to test that methods invoked via the event gateway can trigger even more events.
	//
	// ON OtherService.ChainTwo
	ChainFour(ctx context.Context, request *OtherRequest) (*OtherResponse, error)

	// ChainFail fires after ChainOne, but should always return an error. This will prevent ChainFailAfter
	// from ever actually running.
	//
	// ON OtherService.ChainOne
	ChainFail(ctx context.Context, request *OtherRequest) (*OtherResponse, error)

	// ChainFailAfter is dependent on a successful call to ChainFail... which always fails. So this NEVER runs.
	//
	// ON OtherService.ChainFail
	ChainFailAfter(ctx context.Context, request *OtherRequest) (*OtherResponse, error)
}

// OtherRequest is a basic payload that partially matches the schema of SampleResponse so
// when we invoke service methods through the event gateway, we can make sure that we
// can get the Text value while ignoring everything else from the original payload.
type OtherRequest struct {
	// UniqueThing is just a field that doesn't exist in any other testing response. This ensures
	// that we can use events to decode the values like 'Text' which are present while ignoring those
	// that are not... quietly.
	UniqueThing bool
	// Text is the result of the previous call's invocation.
	Text string
}

// OtherResponse is a single-value output.
type OtherResponse OtherRequest

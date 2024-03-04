// Code generated by Frodo - DO NOT EDIT.
//
//	Timestamp: Sun, 06 Nov 2022 16:46:37 EST
//	Source:    calc/calc_service.go
//	Generator: https://github.com/bridgekitio/frodo
package calc

import (
	"context"
	"fmt"
	"time"

	"github.com/bridgekitio/frodo/example/basic/calc"
)

// MockCalculatorService allows you to program behaviors into a mock instance of CalculatorService. You supply
// dynamic functions named "XxxFunc" to provide the custom behavior; so if your service has a function
// called 'CreateUser', you supply a function for 'CreateUserFunc'.
//
// You do not need to supply behaviors for every single service function; just the ones you plan to
// test. If you do invoke a function without a programmed behavior, it will just return an error
// with a message indicating that it wasn't implemented.
type MockCalculatorService struct {
	AddFunc    func(context.Context, *calc.AddRequest) (*calc.AddResponse, error)
	DoubleFunc func(context.Context, *calc.DoubleRequest) (*calc.DoubleResponse, error)
	MulFunc    func(context.Context, *calc.MulRequest) (*calc.MulResponse, error)
	SubFunc    func(context.Context, *calc.SubRequest) (*calc.SubResponse, error)

	Calls struct {
		Add    callsCalculatorServiceAdd
		Double callsCalculatorServiceDouble
		Mul    callsCalculatorServiceMul
		Sub    callsCalculatorServiceSub
	}
}

/* ---- CalculatorService.Add Mock Support For  ---- */

func (mock *MockCalculatorService) Add(ctx context.Context, request *calc.AddRequest) (*calc.AddResponse, error) {
	mock.Calls.Add = mock.Calls.Add.invoked(*request)
	if mock.AddFunc == nil {
		return nil, fmt.Errorf("CalculatorService.Add not implemented")
	}
	response, err := mock.AddFunc(ctx, request)
	return response, err
}

type callCalculatorServiceAdd struct {
	Time    time.Time
	Request calc.AddRequest
}

type callsCalculatorServiceAdd []callCalculatorServiceAdd

func (calls callsCalculatorServiceAdd) invoked(request calc.AddRequest) callsCalculatorServiceAdd {
	return append(calls, callCalculatorServiceAdd{Time: time.Now(), Request: request})
}

// Times return the total number of times that Add was invoked with any request arguments.
func (calls callsCalculatorServiceAdd) Times() int {
	return len(calls)
}

// TimesFor return the total number of times that Add was invoked with the specific input. Equality
// is determined using == on this 'request' param and the de-referenced one used in the invocation, so
// we'll only county times for those with structural equality.
func (calls callsCalculatorServiceAdd) TimesFor(request calc.AddRequest) int {
	return calls.TimesMatching(func(actual calc.AddRequest) bool {
		return actual == request
	})
}

// TimesMatching return the total number of times that Add was invoked with any
// input that returns true when fed to your predicate function. It's a way to filter by
// requests that meet some requirement more complex than equality (like TimesFor uses).
func (calls callsCalculatorServiceAdd) TimesMatching(pred func(calc.AddRequest) bool) int {
	count := 0
	for _, call := range calls {
		if pred(call.Request) {
			count++
		}
	}
	return count
}

/* ---- CalculatorService.Double Mock Support For  ---- */

func (mock *MockCalculatorService) Double(ctx context.Context, request *calc.DoubleRequest) (*calc.DoubleResponse, error) {
	mock.Calls.Double = mock.Calls.Double.invoked(*request)
	if mock.DoubleFunc == nil {
		return nil, fmt.Errorf("CalculatorService.Double not implemented")
	}
	response, err := mock.DoubleFunc(ctx, request)
	return response, err
}

type callCalculatorServiceDouble struct {
	Time    time.Time
	Request calc.DoubleRequest
}

type callsCalculatorServiceDouble []callCalculatorServiceDouble

func (calls callsCalculatorServiceDouble) invoked(request calc.DoubleRequest) callsCalculatorServiceDouble {
	return append(calls, callCalculatorServiceDouble{Time: time.Now(), Request: request})
}

// Times return the total number of times that Double was invoked with any request arguments.
func (calls callsCalculatorServiceDouble) Times() int {
	return len(calls)
}

// TimesFor return the total number of times that Double was invoked with the specific input. Equality
// is determined using == on this 'request' param and the de-referenced one used in the invocation, so
// we'll only county times for those with structural equality.
func (calls callsCalculatorServiceDouble) TimesFor(request calc.DoubleRequest) int {
	return calls.TimesMatching(func(actual calc.DoubleRequest) bool {
		return actual == request
	})
}

// TimesMatching return the total number of times that Double was invoked with any
// input that returns true when fed to your predicate function. It's a way to filter by
// requests that meet some requirement more complex than equality (like TimesFor uses).
func (calls callsCalculatorServiceDouble) TimesMatching(pred func(calc.DoubleRequest) bool) int {
	count := 0
	for _, call := range calls {
		if pred(call.Request) {
			count++
		}
	}
	return count
}

/* ---- CalculatorService.Mul Mock Support For  ---- */

func (mock *MockCalculatorService) Mul(ctx context.Context, request *calc.MulRequest) (*calc.MulResponse, error) {
	mock.Calls.Mul = mock.Calls.Mul.invoked(*request)
	if mock.MulFunc == nil {
		return nil, fmt.Errorf("CalculatorService.Mul not implemented")
	}
	response, err := mock.MulFunc(ctx, request)
	return response, err
}

type callCalculatorServiceMul struct {
	Time    time.Time
	Request calc.MulRequest
}

type callsCalculatorServiceMul []callCalculatorServiceMul

func (calls callsCalculatorServiceMul) invoked(request calc.MulRequest) callsCalculatorServiceMul {
	return append(calls, callCalculatorServiceMul{Time: time.Now(), Request: request})
}

// Times return the total number of times that Mul was invoked with any request arguments.
func (calls callsCalculatorServiceMul) Times() int {
	return len(calls)
}

// TimesFor return the total number of times that Mul was invoked with the specific input. Equality
// is determined using == on this 'request' param and the de-referenced one used in the invocation, so
// we'll only county times for those with structural equality.
func (calls callsCalculatorServiceMul) TimesFor(request calc.MulRequest) int {
	return calls.TimesMatching(func(actual calc.MulRequest) bool {
		return actual == request
	})
}

// TimesMatching return the total number of times that Mul was invoked with any
// input that returns true when fed to your predicate function. It's a way to filter by
// requests that meet some requirement more complex than equality (like TimesFor uses).
func (calls callsCalculatorServiceMul) TimesMatching(pred func(calc.MulRequest) bool) int {
	count := 0
	for _, call := range calls {
		if pred(call.Request) {
			count++
		}
	}
	return count
}

/* ---- CalculatorService.Sub Mock Support For  ---- */

func (mock *MockCalculatorService) Sub(ctx context.Context, request *calc.SubRequest) (*calc.SubResponse, error) {
	mock.Calls.Sub = mock.Calls.Sub.invoked(*request)
	if mock.SubFunc == nil {
		return nil, fmt.Errorf("CalculatorService.Sub not implemented")
	}
	response, err := mock.SubFunc(ctx, request)
	return response, err
}

type callCalculatorServiceSub struct {
	Time    time.Time
	Request calc.SubRequest
}

type callsCalculatorServiceSub []callCalculatorServiceSub

func (calls callsCalculatorServiceSub) invoked(request calc.SubRequest) callsCalculatorServiceSub {
	return append(calls, callCalculatorServiceSub{Time: time.Now(), Request: request})
}

// Times return the total number of times that Sub was invoked with any request arguments.
func (calls callsCalculatorServiceSub) Times() int {
	return len(calls)
}

// TimesFor return the total number of times that Sub was invoked with the specific input. Equality
// is determined using == on this 'request' param and the de-referenced one used in the invocation, so
// we'll only county times for those with structural equality.
func (calls callsCalculatorServiceSub) TimesFor(request calc.SubRequest) int {
	return calls.TimesMatching(func(actual calc.SubRequest) bool {
		return actual == request
	})
}

// TimesMatching return the total number of times that Sub was invoked with any
// input that returns true when fed to your predicate function. It's a way to filter by
// requests that meet some requirement more complex than equality (like TimesFor uses).
func (calls callsCalculatorServiceSub) TimesMatching(pred func(calc.SubRequest) bool) int {
	count := 0
	for _, call := range calls {
		if pred(call.Request) {
			count++
		}
	}
	return count
}
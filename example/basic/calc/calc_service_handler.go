package calc

import (
	"context"
	"fmt"
)

type CalculatorServiceHandler struct{}

func (svc CalculatorServiceHandler) Add(_ context.Context, req *AddRequest) (*AddResponse, error) {
	fmt.Printf("[Calculator.Add] %v + %v\n", req.A, req.B)
	return &AddResponse{Value: req.A + req.B}, nil
}

func (svc CalculatorServiceHandler) Sub(_ context.Context, req *SubRequest) (*SubResponse, error) {
	fmt.Printf("[Calculator.Sub] %v - %v\n", req.A, req.B)
	return &SubResponse{Value: req.A - req.B}, nil
}

func (svc CalculatorServiceHandler) Mul(_ context.Context, req *MulRequest) (*MulResponse, error) {
	fmt.Printf("[Calculator.Mul] %v * %v\n", req.A, req.B)
	return &MulResponse{Value: req.A * req.B}, nil
}

func (svc CalculatorServiceHandler) Double(_ context.Context, req *DoubleRequest) (*DoubleResponse, error) {
	fmt.Printf("[Calculator.Double] %v\n", req.Value)
	return &DoubleResponse{Value: req.Value * 2}, nil
}

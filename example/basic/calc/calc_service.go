package calc

import (
	"context"
)

// CalculatorService provides the ability to perform basic arithmetic on two numbers.
type CalculatorService interface {
	// Add calculates and returns the sum of two numbers.
	//
	// HTTP 200
	// GET  /add/{A}/{B}
	Add(context.Context, *AddRequest) (*AddResponse, error)

	// Sub calculates and returns the difference between two numbers.
	//
	// HTTP 200
	// GET  /sub/{A}/{B}
	Sub(context.Context, *SubRequest) (*SubResponse, error)

	// Mul calculates and returns the product of two numbers.
	//
	// HTTP 200
	// GET  /multiply/{A}/{B}
	Mul(context.Context, *MulRequest) (*MulResponse, error)

	// Double multiplies the value by 2
	//
	// HTTP 200
	// POST /double/{Value}
	// ON   CalculatorService.Mul
	Double(context.Context, *DoubleRequest) (*DoubleResponse, error)
}

// AddRequest wrangles the two integers you plan to add together.
type AddRequest struct {
	// A is the first number to add.
	A int
	// B is the other number to add.
	B int
}

// AddResponse represents the result of an Add() operation.
type AddResponse struct {
	// Value contains the resulting sum.
	Value int
}

// SubRequest contains the two numbers whose difference you're calculating in Sub().
type SubRequest struct {
	// A is the "minuend" in the subtraction operation.
	A int
	// B is the "subtrahend" in the subtraction operation.
	B int
}

// SubResponse represents the result of an Add() operation.
type SubResponse struct {
	// Value contains the resulting difference.
	Value int
}

// MulRequest contains the two numbers whose product you're calculating in Mul().
type MulRequest struct {
	// A is the first factor to multiply.
	A int
	// B is the other factor to multiply.
	B int
}

// MulResponse represents the result of a Mul() operation.
type MulResponse struct {
	// Value contains the resulting product.
	Value int
}

// DoubleRequest contains the number we're multiplying by 2 in a Double() operation.
type DoubleRequest struct {
	// Value is the number to multiply by 2.
	Value int
}

// DoubleResponse represents the result of a Mul() operation.
type DoubleResponse struct {
	// Value contains the resulting product.
	Value int
}

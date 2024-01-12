package main

import (
	"context"
	"fmt"

	"github.com/bridgekitio/frodo/example/basic/calc"
	calcgen "github.com/bridgekitio/frodo/example/basic/calc/gen"
)

func main() {
	ctx := context.Background()
	client := calcgen.CalculatorServiceClient("localhost:8080")
	addRes, err := client.Add(ctx, &calc.AddRequest{A: 2, B: 5})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Add 2+5=%d\n", addRes.Value)

	mulRes, err := client.Mul(ctx, &calc.MulRequest{A: 6, B: 5})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Mul 6*5=%d\n", mulRes.Value)
}

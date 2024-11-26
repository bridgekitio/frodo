package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bridgekit-io/frodo/example/basic/calc"
	gen "github.com/bridgekit-io/frodo/example/basic/calc/gen"
	"github.com/bridgekit-io/frodo/services"
	"github.com/bridgekit-io/frodo/services/gateways/apis"
	"github.com/bridgekit-io/frodo/services/gateways/events"
)

func main() {
	fmt.Println("Initializing server")

	calcHandler := calc.CalculatorServiceHandler{}
	calcServer := gen.CalculatorServiceServer(calcHandler)

	server := services.NewServer(
		services.Listen(apis.NewGateway(":8080")),
		services.Listen(events.NewGateway()),
		services.Register(calcServer),
	)

	fmt.Println("Server running on http://localhost:8080")
	go server.ShutdownOnInterrupt(2 * time.Second)
	if err := server.Run(context.Background()); err != nil {
		panic(err)
	}
}

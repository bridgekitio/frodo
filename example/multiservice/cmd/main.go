package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bridgekit-io/frodo/example/multiservice/dismissal"
	dismissgateway "github.com/bridgekit-io/frodo/example/multiservice/dismissal/gen"
	"github.com/bridgekit-io/frodo/example/multiservice/greetings"
	greetergateway "github.com/bridgekit-io/frodo/example/multiservice/greetings/gen"
	"github.com/bridgekit-io/frodo/services"
	"github.com/bridgekit-io/frodo/services/gateways/apis"
)

func main() {
	fmt.Println("Initializing server")
	greeterService := greetings.GreeterServiceHandler{}
	dismissService := dismissal.DismissServiceHandler{}

	// Both service APIs will run in a single HTTP server running on localhost:8080.
	server := services.NewServer(
		services.Listen(apis.NewGateway(":8080")),
		services.Register(
			greetergateway.NewGreeterService(greeterService),
			dismissgateway.NewDismissService(dismissService),
		),
	)

	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("Quick examples:")
	fmt.Println("  curl -XPOST -d '{\"Name\":\"Dude\"}' http://localhost:8080/GreeterService.Greet")
	fmt.Println("  curl -XPOST -d '{\"Name\":\"Walter\"}' http://localhost:8080/DismissService.Dismiss")

	// Fire up the API and shut down gracefully when we receive a SIGINT or SIGTERM signal.
	go server.ShutdownOnInterrupt(10 * time.Second)
	if err := server.Run(context.Background()); err != nil {
		panic(err)
	}

	fmt.Println("Bye bye...")
}

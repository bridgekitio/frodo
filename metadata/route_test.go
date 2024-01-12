//go:build unit

package metadata_test

import (
	"context"
	"testing"

	"github.com/bridgekitio/frodo/metadata"
	"github.com/stretchr/testify/suite"
)

func TestRouteSuite(t *testing.T) {
	suite.Run(t, new(RouteSuite))
}

type RouteSuite struct {
	suite.Suite
}

func (suite *RouteSuite) TestDefaults() {
	suite.Equal(metadata.EndpointRoute{}, metadata.Route(nil))
	suite.Equal(metadata.EndpointRoute{}, metadata.Route(context.Background()))
	suite.Nil(metadata.WithRoute(nil, metadata.EndpointRoute{}))
}

func (suite *RouteSuite) TestWithRoute() {
	ctx := context.Background()
	route := metadata.EndpointRoute{}

	route = metadata.EndpointRoute{
		ServiceName: "A",
		Name:        "B",
		Type:        "C",
		Method:      "D",
		Path:        "E",
	}
	ctx = metadata.WithRoute(ctx, route)
	suite.Equal(route, metadata.Route(ctx))

	route = metadata.EndpointRoute{
		ServiceName: "1",
		Name:        "2",
		Type:        "3",
		Method:      "4",
		Path:        "5",
		Status:      204,
	}
	ctx = metadata.WithRoute(ctx, route)
	suite.Equal(route, metadata.Route(ctx))
}

func (suite *RouteSuite) TestRoute_QualifiedName() {
	route := metadata.EndpointRoute{Method: "Does", Path: "Not", Type: "Matter"}
	suite.Equal("", route.QualifiedName())

	route.ServiceName = "Foo"
	route.Name = ""
	suite.Equal("Foo", route.QualifiedName())

	route.ServiceName = ""
	route.Name = "Bar"
	suite.Equal("Bar", route.QualifiedName())

	route.ServiceName = "Foo"
	route.Name = "Bar"
	suite.Equal("Foo.Bar", route.QualifiedName())

	route.ServiceName = "Tasty"
	route.Name = "Beer🍺"
	suite.Equal("Tasty.Beer🍺", route.QualifiedName())
}

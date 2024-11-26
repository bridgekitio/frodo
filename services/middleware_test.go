//go:build unit

package services_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/bridgekit-io/frodo/internal/testext"
	"github.com/bridgekit-io/frodo/services"
	"github.com/stretchr/testify/suite"
)

func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareSuite))
}

type MiddlewareSuite struct {
	suite.Suite
}

func (suite *MiddlewareSuite) TestPipeline_none() {
	results := &testext.Sequence{}
	handler := services.MiddlewareFuncs{}.Then(func(ctx context.Context, req any) (any, error) {
		results.Append(fmt.Sprintf("%v", req))
		return req, nil
	})

	res, err := handler(context.Background(), "Hello")
	suite.NoError(err)
	suite.EqualValues("Hello", res)
	suite.Equal([]string{"Hello"}, results.Values(), "Invalid execution order")
}

func (suite *MiddlewareSuite) TestPipeline_order() {
	results := &testext.Sequence{}
	middleware := services.MiddlewareFuncs{
		func(ctx context.Context, req any, next services.HandlerFunc) (any, error) {
			results.Append("A:Before")
			res, err := next(ctx, req)
			results.Append("A:After")
			return res, err
		},
		func(ctx context.Context, req any, next services.HandlerFunc) (any, error) {
			results.Append("B:Before")
			res, err := next(ctx, req)
			results.Append("B:After")
			return res, err
		},
	}
	handler := middleware.Then(func(ctx context.Context, req any) (any, error) {
		results.Append(fmt.Sprintf("%v", req))
		return req, nil
	})

	res, err := handler(context.Background(), "Hello")
	suite.NoError(err)
	suite.EqualValues("Hello", res)
	suite.Equal([]string{
		"A:Before",
		"B:Before",
		"Hello",
		"B:After",
		"A:After",
	}, results.Values(), "Invalid execution order")
}

func (suite *MiddlewareSuite) TestPipeline_shortCircuit_ok() {
	middleware := services.MiddlewareFuncs{
		func(ctx context.Context, req any, next services.HandlerFunc) (any, error) {
			return "Johnny 5 is ALIVE!", nil
		},
	}
	handler := middleware.Then(func(ctx context.Context, req any) (any, error) {
		return req, nil
	})

	res, err := handler(context.Background(), "Hello")
	suite.NoError(err)
	suite.EqualValues("Johnny 5 is ALIVE!", res)
}

func (suite *MiddlewareSuite) TestPipeline_shortCircuit_error() {
	middleware := services.MiddlewareFuncs{
		func(ctx context.Context, req any, next services.HandlerFunc) (any, error) {
			return nil, errors.New("no disassemble")
		},
	}
	handler := middleware.Then(func(ctx context.Context, req any) (any, error) {
		return req, nil
	})

	res, err := handler(context.Background(), "Hello")
	suite.Error(err)
	suite.Nil(res)
}

func (suite *MiddlewareSuite) TestPipeline_changeResult() {
	middleware := services.MiddlewareFuncs{
		func(ctx context.Context, req any, next services.HandlerFunc) (any, error) {
			res, err := next(ctx, req)
			switch {
			case res.(string) == "Mr. Lebowski":
				return "Dude", nil // That's what you call me...
			case res.(string) == "Error":
				return nil, errors.New("shut the fuck up donny")
			default:
				return res, err
			}
		},
	}
	handler := middleware.Then(func(ctx context.Context, req any) (any, error) {
		return req, nil
	})

	res, err := handler(context.Background(), "Hello")
	suite.NoError(err)
	suite.EqualValues("Hello", res)

	res, err = handler(context.Background(), "Mr. Lebowski")
	suite.NoError(err)
	suite.EqualValues("Dude", res)

	_, err = handler(context.Background(), "Error")
	suite.Error(err)
	suite.Equal("shut the fuck up donny", err.Error())
}

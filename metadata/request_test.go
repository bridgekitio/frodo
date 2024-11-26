//go:build unit

package metadata_test

import (
	"context"
	"testing"

	"github.com/bridgekit-io/frodo/metadata"
	"github.com/stretchr/testify/suite"
)

func TestRequestSuite(t *testing.T) {
	suite.Run(t, new(RequestSuite))
}

type RequestSuite struct {
	suite.Suite
}

func (suite *RequestSuite) TestDefaults() {
	suite.Equal("", metadata.RequestHeader(nil, "Content-Type"))
	suite.Equal("", metadata.RequestHeader(context.Background(), ""))
	suite.Nil(metadata.WithRequestHeaders(nil, map[string][]string{}))
	suite.NotNil(metadata.WithRequestHeaders(context.Background(), nil))

	ctx := metadata.WithRequestHeaders(context.Background(), map[string][]string{})
	suite.Equal("", metadata.RequestHeader(ctx, "Content-Type"))
	suite.Equal("", metadata.RequestHeader(ctx, "Content-Length"))
}

func (suite *RequestSuite) TestWithRequestHeaders() {
	ctx := context.Background()

	ctx = metadata.WithRequestHeaders(ctx, map[string][]string{
		"Content-Type":   {"image/png"},
		"Content-Length": {"42"},
	})
	suite.Equal("image/png", metadata.RequestHeader(ctx, "Content-Type"))
	suite.Equal("42", metadata.RequestHeader(ctx, "Content-Length"))
	suite.Equal("", metadata.RequestHeader(ctx, "Content-Range"))

	// Calling WithRequestHeaders completely overshadows the previous header value map,
	// so neither of the 2 previous values are available through metadata.RequestHeader().
	ctx = metadata.WithRequestHeaders(ctx, map[string][]string{
		"Content-Type":  {"image/jpg"},
		"Content-Range": {"bytes 0-1024/2048"},
		"Accept":        {"application/json", "text/html;q=0.9", "*/*"},
	})
	suite.Equal("image/jpg", metadata.RequestHeader(ctx, "Content-Type"))
	suite.Equal("bytes 0-1024/2048", metadata.RequestHeader(ctx, "Content-Range"))
	suite.Equal("", metadata.RequestHeader(ctx, "Content-Length"))

	// If there are multiple values, return a comma-space delimited string of them
	suite.Equal("application/json, text/html;q=0.9, */*", metadata.RequestHeader(ctx, "Accept"))
}

//go:build unit

package metadata_test

import (
	"context"
	"testing"

	"github.com/bridgekit-io/frodo/metadata"
	"github.com/stretchr/testify/suite"
)

func TestMetadataSuite(t *testing.T) {
	suite.Run(t, new(MetadataSuite))
}

type MetadataSuite struct {
	suite.Suite
}

func (suite *MetadataSuite) TestEncode_Defaults() {
	var ctx context.Context
	suite.Equal(metadata.EncodedBytes(""), metadata.Encode(ctx))

	ctx = context.Background()
	suite.Equal(metadata.EncodedBytes(""), metadata.Encode(ctx))

	// We only encode stuff specifically added to contexts using 'metadata.WithXXX()'
	ctx = context.WithValue(ctx, "Not Related", 42)
	ctx = context.WithValue(ctx, "Also Not Related", 1024)
	suite.Equal(metadata.EncodedBytes(""), metadata.Encode(ctx))
}

func (suite *MetadataSuite) TestDecode_Defaults() {
	ctx := metadata.Decode(nil, ``)
	suite.NotNil(ctx, "Decoding metadata onto a nil context should automatically start w/ a background context")
	suite.Equal("", metadata.Authorization(ctx))
	suite.Equal("", metadata.TraceID(ctx))

	ctx = metadata.Decode(nil, `garbage data`)
	suite.NotNil(ctx, "Decoding metadata onto a nil context should automatically start w/ a background context")
	suite.Equal("", metadata.Authorization(ctx))
	suite.Equal("", metadata.TraceID(ctx))

	ctx = metadata.Decode(nil, `{"Authorization":"Abide","TraceID":"12345"}`)
	suite.NotNil(ctx, "Decoding metadata onto a nil context should automatically start w/ a background context")
	suite.Equal("Abide", metadata.Authorization(ctx))
	suite.Equal("12345", metadata.TraceID(ctx))

	ctx = metadata.Decode(context.Background(), ``)
	suite.Equal("", metadata.Authorization(ctx))
	suite.Equal("", metadata.TraceID(ctx))

	ctx = metadata.Decode(context.Background(), `garbage data`)
	suite.Equal("", metadata.Authorization(ctx))
	suite.Equal("", metadata.TraceID(ctx))

	ctx = metadata.Decode(context.Background(), `{"Authorization":"Abide","TraceID":"12345"}`)
	suite.Equal("Abide", metadata.Authorization(ctx))
	suite.Equal("12345", metadata.TraceID(ctx))
}

func (suite *MetadataSuite) TestEncodeDecode_ignoreTransients() {
	ctx := context.Background()

	// These should be included in the payload when we encode.
	ctx = metadata.WithAuthorization(ctx, "Abide")
	ctx = metadata.WithTraceID(ctx, "12345")
	ctx = metadata.WithValue(ctx, "Foo", "A")
	ctx = metadata.WithValue(ctx, "Bar", "B")
	ctx = metadata.WithValue(ctx, "Baz", "C")

	// These should NOT come along for the encoding/decoding ride.
	ctx = metadata.WithRoute(ctx, metadata.EndpointRoute{Name: "Do Not Encode Me!"})
	ctx = metadata.WithRequestHeaders(ctx, map[string][]string{
		"Content-Type":   {"image/png"},
		"Content-Length": {"100"},
	})

	var stringValue string

	decoded := metadata.Decode(context.Background(), metadata.Encode(ctx))
	suite.Require().NotNil(decoded, "Decoded context should never be nil")

	// Make sure auth, trace id, and values came along for the ride.
	suite.Equal("Abide", metadata.Authorization(decoded))
	suite.Equal("12345", metadata.TraceID(decoded))
	suite.True(metadata.Value(decoded, "Foo", &stringValue))
	suite.Equal("A", stringValue)
	suite.True(metadata.Value(decoded, "Bar", &stringValue))
	suite.Equal("B", stringValue)
	suite.True(metadata.Value(decoded, "Baz", &stringValue))
	suite.Equal("C", stringValue)

	// Make sure that request/route data did not follow along.
	suite.Equal("", metadata.RequestHeader(decoded, "Content-Type"))
	suite.Equal("", metadata.RequestHeader(decoded, "Content-Length"))
	suite.Equal("", metadata.Route(decoded).Name)
}

// Given that values are handled in a slightly more odd fashion, make sure that Encode/Decode
// works just fine with contexts that only specify Authorization and TraceID.
func (suite *MetadataSuite) TestEncodeDecode_noValues() {
	ctx := context.Background()
	ctx = metadata.WithAuthorization(ctx, "Abide")
	ctx = metadata.WithTraceID(ctx, "12345")

	var stringValue string

	decoded := metadata.Decode(context.Background(), metadata.Encode(ctx))
	suite.Require().NotNil(decoded, "Decoded context should never be nil")
	suite.Equal("Abide", metadata.Authorization(decoded))
	suite.Equal("12345", metadata.TraceID(decoded))

	// Should behave like there's no value - normal behavior; no panics.
	suite.False(metadata.Value(decoded, "Foo", &stringValue))
	suite.Equal("", stringValue)
}

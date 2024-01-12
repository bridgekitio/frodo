//go:build unit

package metadata_test

import (
	"context"
	"testing"

	"github.com/bridgekitio/frodo/metadata"
	"github.com/stretchr/testify/suite"
)

func TestTraceIDSuite(t *testing.T) {
	suite.Run(t, new(TraceIDSuite))
}

type TraceIDSuite struct {
	suite.Suite
}

func (suite *TraceIDSuite) TestDefaults() {
	suite.Equal("", metadata.TraceID(nil))
	suite.Equal("", metadata.TraceID(context.Background()))
	suite.Nil(metadata.WithTraceID(nil, ""))
}

func (suite *TraceIDSuite) TestWithTraceID() {
	ctx := context.Background()

	ctx = metadata.WithTraceID(ctx, "Abide")
	suite.Equal("Abide", metadata.TraceID(ctx))

	ctx = metadata.WithTraceID(ctx, "")
	suite.Equal("", metadata.TraceID(ctx))

	ctx = metadata.WithTraceID(ctx, "Abide Harder")
	suite.Equal("Abide Harder", metadata.TraceID(ctx))
}

func (suite *TraceIDSuite) TestNewTraceID() {
	id1 := metadata.NewTraceID()
	id2 := metadata.NewTraceID()
	id3 := metadata.NewTraceID()

	suite.Len(id1, 24, "Generated Trace IDs should be 24 runes long.")
	suite.Len(id2, 24, "Generated Trace IDs should be 24 runes long.")
	suite.Len(id3, 24, "Generated Trace IDs should be 24 runes long.")

	suite.NotEqualf(id1, id2, "Generated Trace IDs should be unique.")
	suite.NotEqualf(id2, id3, "Generated Trace IDs should be unique.")
	suite.NotEqualf(id1, id3, "Generated Trace IDs should be unique.")

	ctx := context.Background()

	ctx = metadata.WithTraceID(ctx, id3)
	suite.Equal(id3, metadata.TraceID(ctx))

	ctx = metadata.WithTraceID(ctx, id2)
	suite.Equal(id2, metadata.TraceID(ctx))

	ctx = metadata.WithTraceID(ctx, id1)
	suite.Equal(id1, metadata.TraceID(ctx))
}

//go:build unit

package metadata_test

import (
	"context"
	"testing"

	"github.com/bridgekitio/frodo/metadata"
	"github.com/stretchr/testify/suite"
)

func TestValuesSuite(t *testing.T) {
	suite.Run(t, new(ValuesSuite))
}

type ValuesSuite struct {
	suite.Suite
}

func (suite *ValuesSuite) TestDefaults() {
	var ctx context.Context
	var intValue int
	suite.Equal(false, metadata.Value(nil, "A", &intValue))
	suite.Equal(false, metadata.Value(context.Background(), "", &intValue))
	suite.Equal(false, metadata.Value(context.Background(), "A", nil))

	ctx = nil
	suite.Nil(metadata.WithValue(nil, "A", 42))

	ctx = metadata.WithValue(context.Background(), "", 42)
	suite.Require().NotNil(ctx)
	suite.Equal(false, metadata.Value(ctx, "", &intValue))

	ctx = metadata.WithValue(context.Background(), "A", nil)
	suite.Require().NotNil(ctx)
	suite.Equal(false, metadata.Value(ctx, "", &intValue))
}

func (suite *ValuesSuite) TestTypes() {
	type Dude struct {
		Bowling       bool
		WhiteRussians int
	}

	ctx := context.Background()
	ctx = metadata.WithValue(ctx, "Int", 42)
	ctx = metadata.WithValue(ctx, "Int8", int8(127))
	ctx = metadata.WithValue(ctx, "String", "Abide")
	ctx = metadata.WithValue(ctx, "Bool", true)
	ctx = metadata.WithValue(ctx, "Float64", 3.14)
	ctx = metadata.WithValue(ctx, "Dude", Dude{Bowling: true, WhiteRussians: 4})
	ctx = metadata.WithValue(ctx, "DudePointer", &Dude{Bowling: false, WhiteRussians: 9})
	ctx = metadata.WithValue(ctx, "Slice", []string{"Walter", "Donny"})
	ctx = metadata.WithValue(ctx, "SlicePointer", &[]string{"Dude", "Maude"})
	ctx = metadata.WithValue(ctx, "🍺", "Yummy")

	var ok bool
	var intValue int
	var int8Value int8
	var stringValue string
	var boolValue bool
	var float64Value float64
	var dudeValue Dude
	var dudePointer *Dude
	var sliceValue []string
	var slicePointer *[]string
	var beerValue string

	ok = metadata.Value(ctx, "Int", &intValue)
	suite.True(ok, "Should be able to fetch int value from metadata")
	suite.Equal(42, intValue)

	ok = metadata.Value(ctx, "Int8", &int8Value)
	suite.True(ok, "Should be able to fetch int8 value from metadata")
	suite.Equal(int8(127), int8Value)

	ok = metadata.Value(ctx, "String", &stringValue)
	suite.True(ok, "Should be able to fetch string value from metadata")
	suite.Equal("Abide", stringValue)

	ok = metadata.Value(ctx, "Bool", &boolValue)
	suite.True(ok, "Should be able to fetch bool value from metadata")
	suite.Equal(true, boolValue)

	ok = metadata.Value(ctx, "Float64", &float64Value)
	suite.True(ok, "Should be able to fetch float64 value from metadata")
	suite.Equal(3.14, float64Value)

	ok = metadata.Value(ctx, "Dude", &dudeValue)
	suite.True(ok, "Should be able to fetch struct value from metadata")
	suite.Equal(true, dudeValue.Bowling)
	suite.Equal(4, dudeValue.WhiteRussians)

	ok = metadata.Value(ctx, "DudePointer", &dudePointer)
	suite.True(ok, "Should be able to fetch struct Pointer value from metadata")
	suite.Require().NotNil(dudePointer)
	suite.Equal(false, dudePointer.Bowling)
	suite.Equal(9, dudePointer.WhiteRussians)

	ok = metadata.Value(ctx, "Slice", &sliceValue)
	suite.True(ok, "Should be able to fetch slice value from metadata")
	suite.Equal([]string{"Walter", "Donny"}, sliceValue)

	ok = metadata.Value(ctx, "SlicePointer", &slicePointer)
	suite.True(ok, "Should be able to fetch slice pointer from metadata")
	suite.Require().NotNil(slicePointer)
	suite.Equal([]string{"Dude", "Maude"}, *slicePointer)

	ok = metadata.Value(ctx, "🍺", &beerValue)
	suite.True(ok, "Should be able to fetch value from metadata using emoji in key")
	suite.Equal("Yummy", beerValue)
}

// For now, we fully expect that if you modify values in one context, it's available
// in any sub-contexts. This is how we make it so that you can set meta values in your
// service handlers, and they'll be available in RPC/Event calls to other services
// even though they might be fired using outer contexts.
func (suite *ValuesSuite) TestMutable() {
	base := metadata.WithValue(context.Background(), "Foo", "A")

	// We're explicitly not keeping these inner contexts b/c the under-the-hood
	// map of values on the 'base' one should be getting modified.
	metadata.WithValue(base, "Bar", "B")
	metadata.WithValue(base, "Baz", "C")
	metadata.WithValue(base, "Goo", "D")

	var value string

	suite.Require().True(metadata.Value(base, "Foo", &value))
	suite.Equal("A", value)

	suite.Require().True(metadata.Value(base, "Bar", &value))
	suite.Equal("B", value)

	suite.Require().True(metadata.Value(base, "Baz", &value))
	suite.Equal("C", value)

	suite.Require().True(metadata.Value(base, "Goo", &value))
	suite.Equal("D", value)

	suite.Require().False(metadata.Value(base, "Nope...", &value))
}

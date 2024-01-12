//go:build unit

package eventsource_test

import (
	"testing"

	"github.com/bridgekitio/frodo/eventsource"
	"github.com/stretchr/testify/assert"
)

func TestNamespace(t *testing.T) {
	assert.Equal(t, "", eventsource.Namespace(""))
	assert.Equal(t, "", eventsource.Namespace("Foo"))
	assert.Equal(t, "Foo", eventsource.Namespace("Foo."))
	assert.Equal(t, "Foo", eventsource.Namespace("Foo.."))
	assert.Equal(t, "", eventsource.Namespace(".Foo"))
	assert.Equal(t, "", eventsource.Namespace("..Foo"))
	assert.Equal(t, "Foo", eventsource.Namespace("Foo.Bar"))
	assert.Equal(t, "Foo", eventsource.Namespace("Foo.Bar.Baz"))
	assert.Equal(t, "Foo", eventsource.Namespace("Foo...Bar.Baz"))
	assert.Equal(t, "🍺", eventsource.Namespace("🍺.Guzzled"))

}

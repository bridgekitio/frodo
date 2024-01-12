//go:build unit

package services_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/bridgekitio/frodo/services"
	"github.com/stretchr/testify/require"
)

func TestStreamResponse_Default(t *testing.T) {
	assert := require.New(t)

	stream := services.StreamResponse{}
	assert.Equal("", readTextStream(stream.Content()))
	assert.Equal("", stream.ContentType())
	assert.Equal(0, stream.ContentLength())

	start, end, size := stream.ContentRange()
	assert.Equal(0, start)
	assert.Equal(0, end)
	assert.Equal(0, size)
}

func TestStreamResponse_Content(t *testing.T) {
	assert := require.New(t)
	stream := services.StreamResponse{}

	stream.SetContent(newTextStream(""))
	assert.Equal("", readTextStream(stream.Content()))

	stream.SetContent(newTextStream("Abide"))
	assert.Equal("Abide", readTextStream(stream.Content()))
}

func TestStreamResponse_ContentType(t *testing.T) {
	assert := require.New(t)
	stream := services.StreamResponse{}

	stream.SetContentType("")
	assert.Equal("", stream.ContentType())

	stream.SetContentType("Abide")
	assert.Equal("Abide", stream.ContentType())
}

func TestStreamResponse_ContentRange(t *testing.T) {
	assert := require.New(t)
	stream := services.StreamResponse{}

	check := func(expectedStart, expectedEnd, expectedSize int) {
		start, end, size := stream.ContentRange()
		assert.Equal(expectedStart, start)
		assert.Equal(expectedEnd, end)
		assert.Equal(expectedSize, size)
	}

	stream.SetContentRange(0, 0, 0)
	check(0, 0, 0)

	// It still allows invalid values - it's up to the gateway/client to deal w/ bad values
	// since they only work with ANY interface implementation of these functions. So even if
	// we make StreamResponse smarter, we still need to support BYO content/range implementations.
	stream.SetContentRange(42, 0, 0)
	check(42, 0, 0)
	stream.SetContentRange(0, 42, 0)
	check(0, 42, 0)
	stream.SetContentRange(0, 0, 42)
	check(0, 0, 42)
	stream.SetContentRange(0, 42, -42)
	check(0, 42, -42)

	stream.SetContentRange(0, 1000, 2048)
	check(0, 1000, 2048)
	stream.SetContentRange(1001, 2000, 2048)
	check(1001, 2000, 2048)
}

func TestStreamResponse_Standard(t *testing.T) {
	assert := require.New(t)
	stream := services.StreamResponse{}
	stream.SetContent(newTextStream("Jeff Lebowski"))
	stream.SetContentType("text/plain")
	stream.SetContentLength(13)

	assert.Equal("Jeff Lebowski", readTextStream(stream.Content()))
	assert.Equal("text/plain", stream.ContentType())
	assert.Equal(13, stream.ContentLength())

	start, end, size := stream.ContentRange()
	assert.Equal(0, start)
	assert.Equal(0, end)
	assert.Equal(0, size)
}

func TestStreamResponse_Chunked(t *testing.T) {
	assert := require.New(t)
	stream := services.StreamResponse{}
	stream.SetContent(newTextStream("Jeff Lebowski"))
	stream.SetContentType("text/plain")

	assert.Equal("Jeff Lebowski", readTextStream(stream.Content()))
	assert.Equal("text/plain", stream.ContentType())
	assert.Equal(0, stream.ContentLength())

	start, end, size := stream.ContentRange()
	assert.Equal(0, start)
	assert.Equal(0, end)
	assert.Equal(0, size)
}

func TestStreamResponse_Ranged(t *testing.T) {
	assert := require.New(t)
	stream := services.StreamResponse{}
	stream.SetContent(newTextStream("Jeff Lebowski"))
	stream.SetContentType("text/plain")
	stream.SetContentRange(500, 1000, 2048)

	assert.Equal("Jeff Lebowski", readTextStream(stream.Content()))
	assert.Equal("text/plain", stream.ContentType())
	assert.Equal(0, stream.ContentLength())

	start, end, size := stream.ContentRange()
	assert.Equal(500, start)
	assert.Equal(1000, end)
	assert.Equal(2048, size)
}

// There's no requirement that content length must match the size attribute of your
// range. The gateway will choose appropriate values based on what the streamed response
// reports to it. Again, since response structs just need to implement the interface, it
// doesn't do us much good to do any error checking here because we're just going to
// re-implement the logic in the gateway.
func TestStreamResponse_RangedWithLength(t *testing.T) {
	assert := require.New(t)
	stream := services.StreamResponse{}
	stream.SetContent(newTextStream("Jeff Lebowski"))
	stream.SetContentType("text/plain")
	stream.SetContentRange(500, 1000, 2048)
	stream.SetContentLength(500)

	assert.Equal("Jeff Lebowski", readTextStream(stream.Content()))
	assert.Equal("text/plain", stream.ContentType())
	assert.Equal(500, stream.ContentLength())

	start, end, size := stream.ContentRange()
	assert.Equal(500, start)
	assert.Equal(1000, end)
	assert.Equal(2048, size)
}

func newTextStream(value string) io.ReadCloser {
	return io.NopCloser(bytes.NewBufferString(value))
}

func readTextStream(stream io.ReadCloser) string {
	if stream == nil {
		return ""
	}
	data, _ := io.ReadAll(stream)
	return string(data)
}

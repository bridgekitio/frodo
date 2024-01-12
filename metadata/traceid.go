package metadata

import (
	"context"
	"math/rand"
	"time"
)

var traceIDRand = rand.New(rand.NewSource(time.Now().UnixNano()))
var traceIDRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var traceIDLen = 24

type contextKeyTraceID struct{}

// TraceID extracts the special metadata value that provides a consistent identifier
// used to trace requests as they go from service to service.
func TraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(contextKeyTraceID{}).(string); ok {
		return id
	}
	return ""
}

// WithTraceID stores this special tracing metadata value on the request context. Typically,
// you should NOT call this directly. The framework will infer/generate/propagate this
// value automatically.
func WithTraceID(ctx context.Context, id string) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, contextKeyTraceID{}, id)
}

// NewTraceID generates a pseudo-random request id for your context/request if one wasn't
// already provided by the client/caller.
func NewTraceID() string {
	id := make([]rune, traceIDLen)
	for i := range id {
		id[i] = traceIDRunes[traceIDRand.Intn(len(traceIDRunes))]
	}
	return string(id)
}

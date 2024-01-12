package metadata

import (
	"context"
	"net/textproto"
	"strings"
)

type contextKeyRequestHeaders struct{}

// RequestHeader fetches a header from the original gateway request. What this corresponds to
// is completely dependent on the gateway that is serving up the request. For instance, if it
// is an HTTP/API gateway, this will read one of the original http.Request headers. If this is
// an event gateway request, this might not give you anything as we don't currently have a
// notion of event headers (yet?).
//
// THESE VALUES DO NOT FOLLOW YOU if your request makes RPC-style calls or triggers other
// event gateways to fire. It only represents the most recent/current request context.
func RequestHeader(ctx context.Context, name string) string {
	if ctx == nil {
		return ""
	}
	if name == "" {
		return ""
	}
	if headers, ok := ctx.Value(contextKeyRequestHeaders{}).(map[string][]string); ok {
		name = textproto.CanonicalMIMEHeaderKey(name)
		return strings.Join(headers[name], ", ")
	}
	return ""
}

// WithRequestHeaders stores header info from the original gateway request. This would be
// HTTP headers for API gateways, for example. You typically should not call this on your
// own as the framework will do that for you as part of our gateways' standard processing.
func WithRequestHeaders(ctx context.Context, headers map[string][]string) context.Context {
	if ctx == nil {
		return ctx
	}
	if headers == nil {
		return ctx
	}
	// Just to make sure nothing mutates your headers after calling this, we'll make a
	// copy of the map. Additionally, we'll normalize the header names so things like case
	// don't actually matter.
	canonicalHeaders := map[string][]string{}
	for name, value := range headers {
		canonicalHeaders[textproto.CanonicalMIMEHeaderKey(name)] = value
	}
	return context.WithValue(ctx, contextKeyRequestHeaders{}, canonicalHeaders)
}

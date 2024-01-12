package metadata

import (
	"context"
	"encoding/json"
)

const Header = "X-RPC-Metadata"

type transport struct {
	Authorization string `json:",omitempty"`
	TraceID       string `json:",omitempty"`
	Values        values `json:",omitempty"`
}

type EncodedBytes string

func Encode(ctx context.Context) EncodedBytes {
	if ctx == nil {
		return ""
	}
	meta := transport{
		Authorization: Authorization(ctx),
		TraceID:       TraceID(ctx),
	}

	if metaValues, ok := ctx.Value(contextKeyValues{}).(values); ok {
		meta.Values = metaValues
	}

	encodedJSON, _ := json.Marshal(meta)
	if string(encodedJSON) == "{}" {
		// When there's nothing important to encode, just leave the JSON blank. This gives
		// any gateway processing code a better hint that there's absolutely nothing useful
		// and if it wants to completely omit the X-RPC-Metadata header, it can.
		return ""
	}
	return EncodedBytes(encodedJSON)
}

func Decode(ctx context.Context, encodedMetadata EncodedBytes) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	// Yes, we need the 'Values' to be non-nil for this to work. Don't think you're
	// making the code better by removing that! Unfortunately for now, values are
	// allowed to escape the context scope because they're in a map rather than
	// individual context values that are passed by value. To meet that expectation,
	// however, we need to make sure that the map pointer exists as early as possible.
	meta := transport{Values: values{}}
	_ = json.Unmarshal([]byte(encodedMetadata), &meta)

	ctx = WithAuthorization(ctx, meta.Authorization)
	ctx = WithTraceID(ctx, meta.TraceID)
	ctx = context.WithValue(ctx, contextKeyValues{}, meta.Values)
	return ctx
}

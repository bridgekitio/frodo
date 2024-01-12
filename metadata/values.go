package metadata

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/bridgekitio/frodo/internal/reflection"
)

type contextKeyValues struct{}

// values provides a lookup for all of the request-scoped data that you want to follow
// you as you make RPC calls from service to service (or client to service).
type values map[string]*valuesEntry

// valuesEntry is a single value in our metadata "Values" map. It tracks two different representations
// of the value. The Value is the actual Go primitive/struct that you're storing. The JSON field is the
// marshaled version of Value that we use when sending this value over the wire.
//
// The idea is that the entire Values map is converted to JSON and sent as an X-RPC-Values header to the
// service we're invoking. When that service receives the header, it will keep the raw JSON for a while
// because it doesn't have any type information about the values' underlying types - this mainly due to the fact
// that Go reflection does not support looking up type information given a package and type name like
// many other languages. For example even if we encoded that the JSON was the type "Baz" from the
// package "github.com/foo/bar", we wouldn't be able to get a reflect.Type to actually construct a new Baz{}
// when we needed it.
//
// Due to that limitation, the receiving service won't actually know the types of the values until some
// code looks one of the values up and feeds an empty instance into the 'out' parameter of the Value() call.
// At that point, we'll have a strongly typed Go value in hand, so we'll perform some lazy
// decoding of the JSON back into the Value. Now you'll have the real value available for the duration
// of the call.
type valuesEntry struct {
	// JSON is the representation of the value we'll use temporarily when sending the metadata from
	// service A to service B. We will unmarshal it back into Value once we have type information later.
	JSON string `json:"value"`
	// Value is the actual Go value for this piece of RPC metadata.
	Value any
}

// MarshalJSON encodes the entry as a JSON object that encodes the value so that it can be embedded
// in an X-RPC-Values header.
func (v *valuesEntry) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	buf.WriteString(`{"value":`)
	err := json.NewEncoder(buf).Encode(v.Value)
	if err != nil {
		return nil, err
	}
	buf.WriteString(`}`)
	return buf.Bytes(), nil
}

// UnmarshalJSON receives the JSON for a single metadata entry and creates it. Since we do
// not have type information yet (see docs for valuesEntry above) we can't actually unmarshal
// the raw value. We'll just strip off the right-hand-side of the JSON attribute since that's
// the marshaled raw value. The first time that someone calls metadata.Value() for this key we
// will unmarshal the value for reals.
func (v *valuesEntry) UnmarshalJSON(data []byte) error {
	length := len(data)
	if length < 11 {
		return nil
	}
	v.JSON = string(data[9 : length-1])
	return nil
}

// unmarshal uses the type information from 'out' to reconstitute the entry's JSON into a real value.
func (v *valuesEntry) unmarshal(out any) error {
	err := json.Unmarshal([]byte(v.JSON), out)
	if err != nil {
		return err
	}
	v.Value = out
	v.JSON = ""
	return nil
}

// Value looks up a single piece of metadata on the specified context. The 'key' is the name of
// the value you're looking for and 'out' is a pointer to the value you want us to fill in - the
// mechanics are similar to json.Unmarshal().
func Value(ctx context.Context, key string, out any) bool {
	if ctx == nil {
		return false
	}
	if key == "" {
		return false
	}

	// Make sure that we even *have* scope values on the context first.
	entries, ok := ctx.Value(contextKeyValues{}).(values)
	if !ok {
		return false
	}

	// We have a scope but nothing for this key
	entry, ok := entries[key]
	if !ok {
		return false
	}

	// We have already reconstituted the raw value from the header json, so just assign
	// the value to the "out" pointer.
	if entry.Value != nil {
		return reflection.Assign(entry.Value, out)
	}

	// You are likely on the server side and are attempting to access a value for the first time,
	// so we need to unmarshal the JSON to get the value to the caller.
	err := entry.unmarshal(out)
	if err != nil {
		log.Printf("metadata error: unmarshal '%s': %v", key, err)
		return false
	}
	return true
}

// WithValue stores a key/value pair in the context metadata. It returns a new context that contains
// the metadata map with your value.
func WithValue(ctx context.Context, key string, value any) context.Context {
	if ctx == nil {
		return nil
	}
	if key == "" {
		return ctx
	}

	entries, ok := ctx.Value(contextKeyValues{}).(values)
	if !ok || entries == nil {
		entries = values{}
		ctx = context.WithValue(ctx, contextKeyValues{}, entries)
	}

	// At some point I would love for this to be not rely on side effects and mutating a map. I'd
	// rather that adding new values create copies of the metadata structure, so it's more thread
	// safe like normal context values. But this will work for now...
	entries[key] = &valuesEntry{Value: value}
	return ctx
}

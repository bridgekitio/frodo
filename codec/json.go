package codec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/bridgekitio/frodo/internal/reflection"
)

// JSONEncoder simply uses the standard 'encoding/json' package to convert raw service
// structs into their equivalent JSON.
type JSONEncoder struct{}

// ContentType returns "application/json", the expected MIME content type this encoder handles.
func (encoder JSONEncoder) ContentType() string {
	return "application/json"
}

// Encode uses the standard encoding/json package to encode the value's JSON onto the writer.
func (encoder JSONEncoder) Encode(writer io.Writer, value any) error {
	if writer == nil {
		return fmt.Errorf("json encoder: writer error: nil writer")
	}
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		return fmt.Errorf("json encoder: writer error: %w", err)
	}
	return nil
}

// EncodeValues encodes converts a raw Go object into a list of key/value pairs. For
// instance {"User.ID":"123", "User.ContactInfo.Email":"me@you.com"}.
func (encoder JSONEncoder) EncodeValues(value any) url.Values {
	values := url.Values{}
	encoder.encodeValues("", value, values)
	return values
}

func (encoder JSONEncoder) encodeValues(prefix string, rawValue any, out url.Values) {
	if rawValue == nil {
		return
	}

	valueReflect := reflect.Indirect(reflect.ValueOf(rawValue))
	valueType := valueReflect.Type()

	for i := 0; i < valueType.NumField(); i++ {
		valueField := valueReflect.Field(i)

		// Skip unexported and nil fields.
		if reflection.IsNil(valueField) || !valueField.CanInterface() {
			continue
		}

		field := valueType.Field(i)
		fieldKey := strings.TrimPrefix(prefix+"."+reflection.BindingName(field), ".")
		fieldValue := valueField.Interface()

		// We want to honor your desired JSON formats. The only tweak we make is that we strip
		// the outer quotes if your value marshals to a JSON string. The JSON decoder will automatically
		// wrap string-looking values in quotes, so let the value be the raw text inside it.
		if marshal, ok := fieldValue.(json.Marshaler); ok {
			if data, err := marshal.MarshalJSON(); err == nil {
				out.Set(fieldKey, strings.Trim(string(data), `"`))
				continue
			}
		}

		// Recursively add child attributes to *this* list using a "ParentStruct.ChildStruct.GrandchildStruct"
		// style naming convention so that we can include nested values.
		if reflection.IsStructOrPointerTo(valueField.Type()) {
			encoder.encodeValues(fieldKey, fieldValue, out)
			continue
		}

		// It's some primitive value (string, number, bool, etc.), so output the attribute.
		// Probably doesn't handle map/slice types nicely. Will deal with later.
		switch reflection.IndirectTypeKind(valueField.Type()) {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			out.Set(fieldKey, fmt.Sprintf("%d", valueField.Interface()))
		case reflect.Float32, reflect.Float64:
			out.Set(fieldKey, fmt.Sprintf("%f", valueField.Interface()))
		default:
			out.Set(fieldKey, fmt.Sprintf("%v", valueField.Interface()))
		}
	}
}

// JSONDecoder creates the default codec that uses encoding/json to apply body/path/query data
// to service request models. This unifies the processing so that every source of values is decoded
// using the same semantics. The goal is that whether a value comes in from the body or the path or the
// query string, if you supplied a custom UnmarshalJSON() function for a type, that will work.
//
// It assumes that reader/body data is JSON, naturally. Path and query values, however, get massaged into
// JSON so that we can use the standard library's JSON package to unmarshal that data onto your out value.
// For example, let's assume that we have the following query string:
//
//	?first=Bob&last=Smith&age=39&address.city=Seattle&enabled=true
//
// This decoder will first create 5 separate JSON objects:
//
//	{ "first": "Bob" }
//	{ "last": "Smith" }
//	{ "age": 39 }
//	{ "address": { "city": "Seattle" } }     <-- notice how we handle nested values separated by "."
//	{ "enabled": true }
//
// After generating each value, the decoder will feed the massaged JSON to a 'json.Decoder' and standard
// JSON marshaling rules will overlay each one onto your 'out' value.
type JSONDecoder struct {
	Loose bool
}

// Decode simply uses standard encoding/json to populate your 'out' value w/ JSON from the reader.
func (decoder JSONDecoder) Decode(data io.Reader, out any) error {
	if data == nil || data == http.NoBody {
		return nil
	}
	if err := json.NewDecoder(data).Decode(out); err != nil {
		return fmt.Errorf("json decoder: reader error: %w", err)
	}
	return nil
}

// DecodeValues accepts key/value mappings like "User.ID":"123" and uses JSON-style
// decoding to fill your 'out' value with that data.
func (decoder JSONDecoder) DecodeValues(values url.Values, out any) error {
	if len(values) == 0 {
		return nil
	}

	// To keep the logic more simple (but fast enough for most use cases), we will generate a separate
	// JSON representation of each value and run it through the JSON decoder. To make things a bit more
	// efficient, we'll re-use the buffer/reader and the JSON decoder for each value we unmarshal
	// on the output. This way we only suffer one buffer allocation no matter how many values we handle.
	//
	// For instance, even if the input was "id=123&name=Bob&address.city=Seattle", we won't try to generate
	// a single giant-ass JSON object. We'll make 3 separate ones:
	//
	//     {"id": "123"}
	//     {"name": "Bob"}
	//     {"address": {"city": "Seattle"}}
	//
	// But we won't create 3 separate buffers/decoders. We'll just re-use the same one and reset it after each
	// value to reduce allocations.
	buf := &bytes.Buffer{}
	ctx := jsonBindingContext{
		buf:     buf,
		decoder: json.NewDecoder(buf),
	}

	outValue := reflect.Indirect(reflect.ValueOf(out))

	// Chances are that the input value was from 'endpoint.NewInput()' which returns
	// a value of type 'any' which is an interface. We need to peek under that pointer
	// and get at the juicy struct underneath it.
	if outValue.Kind() == reflect.Interface {
		outValue = reflect.Indirect(outValue.Elem())
	}

	for key, value := range values {
		keySegments := strings.Split(key, ".")

		// Follow the segments of the key and determine the JSON type of the last segment. So if you
		// are binding the key "foo.bar.baz", we'll look at the Go data type of the "baz" field once
		// we've followed the path "out.foo.bar". This will spit back our enum for the JSON data type
		// that will most naturally unmarshal to the Go type. So if the Go data type for the "baz" field
		// is uint16 then we'd expect this to return 'jsonTypeNumber'. If "baz" were a string then
		// we'd expect this to return 'jsonTypeString', and so on.
		valueType := decoder.keyToJSONType(outValue, keySegments, value[0])

		// We didn't find a field path with that name (e.g. the key was "name" but there was no field called "name")
		if valueType == jsonTypeNil {
			continue
		}

		// Maybe you provided "foo.bar.baz=4" and there is a field at "out.foo.bar.baz", but it's
		// a struct of some kind, so "4" is not enough to properly bind it. Arrays/slices we'll handle
		// in a future version... maybe.
		if valueType == jsonTypeArray {
			continue
		}

		// Convert the parameter "foo.bar.baz=4" into {"foo":{"bar":{"baz":4}}} so that the standard
		// JSON decoder can work its magic to apply that to 'out' properly.
		ctx.buf.Reset()
		decoder.writeParamJSON(ctx.buf, keySegments, value[0], valueType)

		// Now that we have a close-enough JSON representation of your parameter, let the standard
		// JSON decoder do its magic.
		switch err := ctx.decoder.Decode(out); {
		case decoder.Loose:
			// Ignore errors decoding individual fields. Set valid fields, ignore bad ones.
		case err == nil:
			// We wrote this field successfully. Move on.
		case err != nil:
			// Not in "Loose" mode, so cause the entire decoding to fail.
			return fmt.Errorf("json decoder: value error: '%s'='%s': %w", key, value[0], err)
		}
	}
	return nil
}

// writeParamJSON accepts the decomposed parameter key (e.g. "foo.bar.baz") and the raw string value (e.g. "moo")
// and writes JSON to the buffer which can be used in standard JSON decoding to apply the value to the out
// object (e.g. `{"foo":{"bar":{"baz":"moo"}}}`).
func (decoder JSONDecoder) writeParamJSON(buf *bytes.Buffer, keySegments []string, value string, valueType jsonType) {
	for _, keySegment := range keySegments {
		buf.WriteString(`{"`)
		buf.WriteString(keySegment)
		buf.WriteString(`":`)
	}
	decoder.writeDecodingValueJSON(buf, value, valueType)
	for i := 0; i < len(keySegments); i++ {
		buf.WriteString("}")
	}
}

// writeDecodingValueJSON outputs the right-hand-side of the JSON we're going to use to try and bind
// this value. For instance, when the binder is creating the JSON {"name":"bob"} for the
// parameter "name=bob", this function determines that "bob" is supposed to be written as a string
// and will write `"bob"` to the buffer.
func (decoder JSONDecoder) writeDecodingValueJSON(buf *bytes.Buffer, value string, valueType jsonType) {
	switch valueType {
	case jsonTypeString:
		// Add quotes around string values if they don't already exist. This way we can support
		// natural values like they'd appear in a URL path or query string -- or we can accept
		// values as they appear if you go straight to/from MarshalJSON/UnmarshalJSON.
		if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			buf.WriteString(value)
		} else {
			buf.WriteString(`"`)
			buf.WriteString(value)
			buf.WriteString(`"`)
		}
	case jsonTypeNumber, jsonTypeBool:
		buf.WriteString(value)
	case jsonTypeObject:
		buf.WriteString(value)
	default:
		// Whether it's a nil (unknown) or object type, the binder doesn't support that type
		// of value, so just write null to avoid binding anything if we can help it.
		buf.WriteString("null")
	}
}

// jsonType describes all possible JSON data types.
type jsonType int

const (
	jsonTypeNil    = jsonType(0)
	jsonTypeString = jsonType(1)
	jsonTypeNumber = jsonType(2)
	jsonTypeBool   = jsonType(3)
	jsonTypeObject = jsonType(4)
	jsonTypeArray  = jsonType(5)
)

// keyToJSONType looks at your parameter key (e.g. "foo.bar.baz") and your value (e.g. "12345"), and
// indicates how we should format the value when creating binding JSON. It will use reflection to
// traverse the Go attributes foo, then bar, then baz, and return the most appropriate JSON type
// for the go type. For instance if "baz" is a uint16, the most appropriate jsonType is jsonTypeNumber.
func (decoder JSONDecoder) keyToJSONType(outValue reflect.Value, key []string, value string) jsonType {
	keyLength := len(key)
	if keyLength < 1 {
		return jsonTypeNil
	}
	if outValue.Kind() != reflect.Struct {
		return jsonTypeNil
	}

	// Follow the path of attributes described by the key, so if the key was "foo.bar.baz" then look up
	// "foo" on the out value, then the "bar" attribute on that type, then the "baz" attribute on that type.
	// Once we exit the loop, 'actualType' should be the type of that nested "baz" field, and we can
	// determine the correct JSON type from there.
	actualType := reflection.FlattenPointerType(outValue.Type())
	for i := 0; i < keyLength; i++ {
		field, ok := reflection.FindField(actualType, key[i])
		if !ok {
			return jsonTypeNil
		}
		actualType = reflection.FlattenPointerType(field.Type)
	}

	// Now that we have the Go type for the field that will ultimately be populated by this parameter/value,
	// we need to do a quick double check. The field's Go type might be a type alias for an int64 so the
	// natural choice for a JSON binding would be to use a number (which is what 't' will resolve to).
	//
	// But... what if the user provided the value "5m2s" for that field? If we blindly treat the value like
	// a number, we'll end up with JSON that looks like {"baz":5m2s} which is invalid. We need to quote that
	// value for it to remain valid JSON. So you only get to be a number/boolean if your parameter's value
	// looks like one of those values, too.
	//
	// The canonical use-case for this situation is if you define a custom type alias like this:
	//
	// type ISODuration int64
	//
	// You then implement the MarshalJSON() and UnmarshalJSON() functions so that it supports ISO duration formats
	// such as "PT3M49S". By looking at the Go type you'd think that the incoming param value should be a
	// JSON number (since the duration is an int64), but the value doesn't "look" like a number; it looks
	// like a freeform string. As a result, we need to build the binding JSON {"foo":"PT3M49S"} since we will
	// treat the right-hand side as a string rather than {"foo":PT3M48S} which is not valid.
	t := decoder.typeToJSONType(actualType)
	switch {
	case t == jsonTypeBool && !decoder.looksLikeBoolJSON(value):
		return jsonTypeString

	case t == jsonTypeNumber && !decoder.looksLikeNumberJSON(value):
		return jsonTypeString

	case t == jsonTypeObject && decoder.looksLikeObjectJSON(value):
		return jsonTypeObject

	case t == jsonTypeObject && decoder.looksLikeNumberJSON(value):
		return jsonTypeNumber

	case t == jsonTypeObject && decoder.looksLikeBoolJSON(value):
		return jsonTypeBool

	case t == jsonTypeObject:
		return jsonTypeString

	default:
		return t
	}
}

// typeToJSONType looks at the Go type of some field on a struct and returns the JSON data type
// that will most likely unmarshal to that field w/o an error.
func (decoder JSONDecoder) typeToJSONType(actualType reflect.Type) jsonType {
	switch actualType.Kind() {
	case reflect.String:
		return jsonTypeString
	case reflect.Bool:
		return jsonTypeBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return jsonTypeNumber
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return jsonTypeNumber
	case reflect.Float32, reflect.Float64:
		return jsonTypeNumber
	case reflect.Array, reflect.Slice:
		return jsonTypeArray
	case reflect.Map, reflect.Struct:
		return jsonTypeObject
	default:
		return jsonTypeNil
	}
}

// looksLikeBoolJSON determines if the raw parameter value looks like a boolean value (i.e. true/false).
func (decoder JSONDecoder) looksLikeBoolJSON(value string) bool {
	value = strings.ToLower(value)
	return value == "true" || value == "false"
}

// looksLikeNumberJSON determines if the raw parameter value looks like it can be formatted as a JSON
// number. Basically, does it only contain digits and a decimal point. Currently, this only supports using
// periods as decimal points. A future iteration might support using /x/text/language to support commas
// as decimals points.
func (decoder JSONDecoder) looksLikeNumberJSON(value string) bool {
	for _, r := range value {
		if r == '.' {
			continue
		}
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (decoder JSONDecoder) looksLikeObjectJSON(value string) bool {
	return strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")
}

// jsonBindingContext carries our buffer/decoder context through all the binding operations so
// that all values can share resources (e.g. binding the path params can piggy-back off of the
// work of binding the query string).
type jsonBindingContext struct {
	buf     *bytes.Buffer
	decoder *json.Decoder
}

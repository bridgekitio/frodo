package codec

import (
	"io"
	"net/url"
)

func New() Registry {
	jsonEncoder := JSONEncoder{}
	jsonDecoder := JSONDecoder{}
	return Registry{
		defaultEncoder: jsonEncoder,
		defaultDecoder: jsonDecoder,
		encoders:       map[string]Encoder{"application/json": jsonEncoder},
		decoders:       map[string]Decoder{"application/json": jsonDecoder},

		defaultValueEncoder: jsonEncoder,
		defaultValueDecoder: jsonDecoder,
		valueEncoders:       map[string]ValueEncoder{"application/json": jsonEncoder},
		valueDecoders:       map[string]ValueDecoder{"application/json": jsonDecoder},
	}
}

// Registry helps you wrangle a collection of encoders/decoders such that you can
// choose specific ones at runtime. For instance, at runtime you can decide if you
// want to use a JSON encoder or an XML one (ew...).
type Registry struct {
	defaultEncoder Encoder
	defaultDecoder Decoder
	encoders       map[string]Encoder
	decoders       map[string]Decoder

	defaultValueEncoder ValueEncoder
	defaultValueDecoder ValueDecoder
	valueEncoders       map[string]ValueEncoder
	valueDecoders       map[string]ValueDecoder
}

// DefaultEncoder returns the encoder that you should use if you have no specific preference.
func (reg Registry) DefaultEncoder() Encoder {
	return reg.defaultEncoder
}

// DefaultValueEncoder returns the ValueEncoder that you should use if you have no specific preference.
func (reg Registry) DefaultValueEncoder() ValueEncoder {
	return reg.defaultValueEncoder
}

// Encoder will return the Encoder for the first content type that we have a valid encoder for.
func (reg Registry) Encoder(contentTypes ...string) Encoder {
	for _, contentType := range contentTypes {
		if encoder, ok := reg.encoders[contentType]; ok {
			return encoder
		}
	}
	return reg.DefaultEncoder()
}

// ValueEncoder will return the ValueEncoder for the first content type that we have a valid encoder for.
func (reg Registry) ValueEncoder(contentTypes ...string) ValueEncoder {
	for _, contentType := range contentTypes {
		if encoder, ok := reg.valueEncoders[contentType]; ok {
			return encoder
		}
	}
	return reg.DefaultValueEncoder()
}

// DefaultDecoder returns the decoder that you should use if you have no specific preference.
func (reg Registry) DefaultDecoder() Decoder {
	return reg.defaultDecoder
}

// DefaultValueDecoder returns the ValueDecoder that you should use if you have no specific preference.
func (reg Registry) DefaultValueDecoder() ValueDecoder {
	return reg.defaultValueDecoder
}

// Decoder will return the Decoder for the first content type that we have a valid decoder for.
func (reg Registry) Decoder(contentTypes ...string) Decoder {
	for _, contentType := range contentTypes {
		if decoder, ok := reg.decoders[contentType]; ok {
			return decoder
		}
	}
	return reg.DefaultDecoder()
}

// ValueDecoder will return the ValueDecoder for the first content type that we have a valid decoder for.
func (reg Registry) ValueDecoder(contentTypes ...string) ValueDecoder {
	for _, contentType := range contentTypes {
		if decoder, ok := reg.valueDecoders[contentType]; ok {
			return decoder
		}
	}
	return reg.DefaultValueDecoder()
}

// Decoder describes a mechanism used to decode streams of JSON/whatever data onto your
// service structs. The DecodeValues operation also lets you apply individual values to
// the struct using similar semantics.
type Decoder interface {
	// Decode will unmarshal the reader data onto the 'out' value.
	Decode(data io.Reader, out any) error
}

// ValueDecoder accepts URL encoded values (e.g. "User.ContactInfo.Email"->"foo@bar.com") and
// can unmarshal them onto the out value of your choice.
type ValueDecoder interface {
	// DecodeValues will unmarshal all of the individual values onto the 'out' value.
	DecodeValues(values url.Values, out any) error
}

// Encoder is used to encode streams of JSON/whatever given raw service structs.
type Encoder interface {
	// ContentType returns the MIME encoding type descriptor for the underlying data type
	// that this encoder expects to work with (e.g. "application/json").
	ContentType() string
	// Encode converts the value into the implementation's target data format and writes
	// the raw bytes to the given writer.
	Encode(writer io.Writer, value any) error
}

// ValueEncoder describes a component capable of taking a raw Go value and turning into a map
// of values such as "User.ContactInfo.Email"->"foo@bar.com".
type ValueEncoder interface {
	// EncodeValues accepts a raw Go value and turns it into a map of individual values
	// such as "User.ContactInfo.Email"->"foo@bar.com".
	EncodeValues(value any) url.Values
}

// NopDecoder satisfies the Decoder interface, but does not actually do any work.
type NopDecoder struct{}

// Decode just returns nil. It does not read anything from your io.Reader, either.
func (NopDecoder) Decode(_ io.Reader, _ any) error {
	return nil
}

// DecodeValues just returns nil to quietly complete.
func (NopDecoder) DecodeValues(_ map[string][]string, _ any) error {
	return nil
}

// NopEncoder satisfies the Encoder interface, but does not actually do any work. It
// doesn't look at your input values and will not write anything to the writers you provide.
type NopEncoder struct{}

// ContentType returns "" to indicate that there's no specific encoding.
func (NopEncoder) ContentType() string {
	return ""
}

// Encode does nothing and writes nothing. It simply returns nil to avoid making noise.
func (NopEncoder) Encode(_ io.Writer, _ any) error {
	return nil
}

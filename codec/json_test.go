//go:build unit

package codec_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bridgekit-io/frodo/codec"
	"github.com/bridgekit-io/frodo/internal/testext"
	"github.com/stretchr/testify/suite"
)

func TestJSONSuite(t *testing.T) {
	suite.Run(t, new(JSONSuite))
}

type JSONSuite struct {
	suite.Suite
}

func (suite *JSONSuite) TestDecode_defaults() {
	decoder := codec.JSONDecoder{}

	msg := "Decoding a nil reader should quietly return w/o error."
	suite.NoError(decoder.Decode(nil, nil), msg)
	suite.NoError(decoder.Decode(nil, &testStruct{}), msg)

	msg = "Decoding an HTTP body w/ no content should quietly return w/o error."
	suite.NoError(decoder.Decode(http.NoBody, nil), msg)
	suite.NoError(decoder.Decode(http.NoBody, &testStruct{}), msg)

	msg = "Decoding an empty reader should behave exactly like standard library (EOF error)."
	suite.Error(decoder.Decode(&bytes.Buffer{}, nil), msg)
	suite.Error(decoder.Decode(&bytes.Buffer{}, &testStruct{}), msg)
}

// Ensure that you get errors when *valid* JSON doesn't match the Go types you provide.
func (suite *JSONSuite) TestDecode_invalidTypes() {
	var intValue int
	var stringValue string
	var structType testStruct

	decoder := codec.JSONDecoder{}
	suite.Error(decoder.Decode(bytes.NewBufferString(`"Hello"`), &intValue))
	suite.Error(decoder.Decode(bytes.NewBufferString(`true`), &intValue))
	suite.Error(decoder.Decode(bytes.NewBufferString(`3.14`), &intValue))
	suite.Error(decoder.Decode(bytes.NewBufferString(`{"IntValue":13}`), &intValue))
	suite.Error(decoder.Decode(bytes.NewBufferString(`[13]`), &intValue))
	suite.Error(decoder.Decode(bytes.NewBufferString(`13`), &stringValue))
	suite.Error(decoder.Decode(bytes.NewBufferString(`"Hello"`), &structType))
	suite.Error(decoder.Decode(bytes.NewBufferString(`"{}"`), &structType)) // this is a JSON string, not an object!

}

// Ensures that we can call Decode and get the right data when the JSON lines up perfectly
// with the Go types you provide.
func (suite *JSONSuite) TestDecode_validTypes() {
	decoder := codec.JSONDecoder{}

	var intValue int
	suite.NoError(decoder.Decode(bytes.NewBufferString("42"), &intValue))
	suite.Equal(42, intValue)

	var floatValue float64
	suite.NoError(decoder.Decode(bytes.NewBufferString("3.14"), &floatValue))
	suite.Equal(3.14, floatValue)

	var stringValue string
	suite.NoError(decoder.Decode(bytes.NewBufferString(`"Frodo"`), &stringValue))
	suite.Equal("Frodo", stringValue)

	var boolValue bool
	suite.NoError(decoder.Decode(bytes.NewBufferString(`true`), &boolValue))
	suite.Equal(true, boolValue)

	var sliceValue []string
	suite.NoError(decoder.Decode(bytes.NewBufferString(`["A", "B", "C"]`), &sliceValue))
	suite.Equal([]string{"A", "B", "C"}, sliceValue)

	// Don't need to build JSON by hand. Just use the standard library to get valid JSON
	// representing this object.
	expected := testStruct{
		String:  "Hello",
		Int:     42,
		Int8:    -1,
		Float64: 3.14,
		Bool:    true,
		User: &testStructUser{
			ID:   "123",
			Name: "The Dude",
			AuditTrail: testStructTimestamp{
				Created:  time.Date(2000, time.February, 28, 10, 44, 22, 0, time.UTC),
				Modified: time.Date(2022, time.September, 3, 14, 22, 12, 33, time.UTC),
			},
		},
		RemappedUser: &testStructUser{
			ID:   "456",
			Name: "Walter",
			AuditTrail: testStructTimestamp{
				Created:  time.Date(2001, time.February, 28, 10, 44, 22, 0, time.UTC),
				Modified: time.Date(2021, time.September, 3, 14, 22, 12, 33, time.UTC),
			},
		},
	}
	inputJSON, _ := json.Marshal(expected)

	var structValue = testStruct{}
	suite.NoError(decoder.Decode(bytes.NewBuffer(inputJSON), &structValue))
	suite.Require().NotNil(structValue.User)
	suite.Require().NotNil(structValue.RemappedUser)
	suite.Equal(*expected.User, *structValue.User)
	suite.Equal(*expected.RemappedUser, *structValue.RemappedUser)
	suite.Equal(expected.String, structValue.String)
	suite.Equal(expected.Int, structValue.Int)
	suite.Equal(expected.Int8, structValue.Int8)
	suite.Equal(expected.Float64, structValue.Float64)
	suite.Equal(expected.Bool, structValue.Bool)
}

func (suite *JSONSuite) TestDecode_remapping() {
	decoder := codec.JSONDecoder{}

	// Name is remapped to 'GoesBy', so this should not apply the Name field.
	var value testStructUser
	suite.NoError(decoder.Decode(bytes.NewBufferString(`{"ID":"Frodo", "Name":"The Dude"}`), &value))
	suite.Equal("Frodo", value.ID)
	suite.Equal("", value.Name)

	suite.NoError(decoder.Decode(bytes.NewBufferString(`{"ID":"Frodo", "goes_by":"The Dude"}`), &value))
	suite.Equal("Frodo", value.ID)
	suite.Equal("The Dude", value.Name)
}

func (suite *JSONSuite) TestEncoder_contentType() {
	encoder := codec.JSONEncoder{}
	suite.Equal("application/json", encoder.ContentType())
}

func (suite *JSONSuite) TestEncoder_validTypes() {
	runTestCase := func(value any) {
		expectedJSON, _ := json.Marshal(value)
		buf := &bytes.Buffer{}
		suite.NoError(codec.JSONEncoder{}.Encode(buf, value))
		suite.Equal(
			strings.TrimSpace(string(expectedJSON)),
			strings.TrimSpace(buf.String()),
		)
	}

	runTestCase("Hello")
	runTestCase(42)
	runTestCase(uint8(42))
	runTestCase(3.14)
	runTestCase(true)
	runTestCase([]int{1, 2, 3, 4})
	runTestCase(testStructUser{
		Name:       "The Dude",
		AuditTrail: testStructTimestamp{Created: time.Date(2022, time.September, 3, 14, 22, 12, 33, time.UTC)},
	})
}

func (suite *JSONSuite) TestEncoder_invalidTypes() {
	runTestCase := func(value any) {
		suite.Error(codec.JSONEncoder{}.Encode(&bytes.Buffer{}, value))
	}

	runTestCase(func() {})
	runTestCase(struct{ Foo chan int }{Foo: make(chan int)})
}

// Make sure that we abide the json:"-" field attribute.
func (suite *JSONSuite) TestEncoder_excludedField() {
	buf := &bytes.Buffer{}
	suite.NoError(codec.JSONEncoder{}.Encode(buf, testStructUser{Ignore: "Foo"}))
	suite.NotContains(buf.String(), "Ignore")
	suite.NotContains(buf.String(), "Foo")
}

func (suite *JSONSuite) TestDecodeValues_defaults() {
	value := testStruct{}
	decoder := codec.JSONDecoder{}

	suite.NoError(decoder.DecodeValues(nil, nil))
	suite.NoError(decoder.DecodeValues(map[string][]string{}, nil))

	suite.NoError(decoder.DecodeValues(nil, &value))
	suite.Equal("", value.String)

	suite.NoError(decoder.DecodeValues(map[string][]string{}, &value))
	suite.Equal("", value.String)
}

func (suite *JSONSuite) TestDecodeValues_validTypes() {
	decoder := codec.JSONDecoder{}

	var value testStruct
	var values = map[string][]string{
		"String":                   {"Hello"},
		"Int":                      {"42"},
		"Int8":                     {"123"},
		"Float64":                  {"3.14", "99.123"}, // only support the first value
		"Bool":                     {"true"},
		"User.ID":                  {"123"},
		"User.goes_by":             {"The Dude"},      // should listen to JSON attributes
		"User.Name":                {"Jeff Lebowski"}, // remapped to 'goes_by' so ignore this
		"alias.ID":                 {"456"},
		"alias.goes_by":            {"Walter"},
		"alias.AuditTrail.Deleted": {"true"},
	}
	suite.NoError(decoder.DecodeValues(values, &value))
	suite.Equal("Hello", value.String)
	suite.Equal(42, value.Int)
	suite.Equal(int8(123), value.Int8)
	suite.Equal(3.14, value.Float64)
	suite.Equal(true, value.Bool)
	suite.Require().NotNil(value.User)
	suite.Equal("123", value.User.ID)
	suite.Equal("The Dude", value.User.Name)
	suite.Equal(false, value.User.AuditTrail.Deleted)
	suite.Require().NotNil(value.RemappedUser)
	suite.Equal("456", value.RemappedUser.ID)
	suite.Equal("Walter", value.RemappedUser.Name)
	suite.Equal(true, value.RemappedUser.AuditTrail.Deleted)
}

func (suite *JSONSuite) TestDecodeValues_invalidTypes() {
	decoder := codec.JSONDecoder{}

	var value testStruct
	var values = map[string][]string{
		"Int": {"Hello"},
		// Even though these are okay, we won't map them b/c "Int" is bad.
		"User.ID":                  {"123"},
		"User.goes_by":             {"The Dude"},
		"alias.ID":                 {"456"},
		"alias.goes_by":            {"Walter"},
		"alias.AuditTrail.Deleted": {"true"},
	}
	suite.Error(decoder.DecodeValues(values, &value))
	suite.Equal(0, value.Int)
}

// Ensures that if we set Loose=true on the decoder that DecodeValues will ignore
// bad fields and just set the ones that work.
func (suite *JSONSuite) TestDecodeValues_loose() {
	decoder := codec.JSONDecoder{Loose: true}

	var value testStruct
	var values = map[string][]string{
		"String":        {"Hello"},
		"Int":           {"42"},
		"Float64":       {"3.14", "99.123"}, // only support the first value
		"Bool":          {"true"},
		"User.ID":       {"123"},
		"User.goes_by":  {"The Dude"}, // should listen to JSON attributes
		"alias.ID":      {"456"},
		"alias.goes_by": {"Walter"},

		// Invalid fields that should be ignored...
		"Int8":                     {"Fart"},
		"alias.AuditTrail.Deleted": {"Nope"},
	}
	suite.NoError(decoder.DecodeValues(values, &value))
	suite.Equal("Hello", value.String)
	suite.Equal(42, value.Int)
	suite.Equal(3.14, value.Float64)
	suite.Equal(true, value.Bool)
	suite.Require().NotNil(value.User)
	suite.Equal("123", value.User.ID)
	suite.Equal("The Dude", value.User.Name)
	suite.Equal(false, value.User.AuditTrail.Deleted)
	suite.Require().NotNil(value.RemappedUser)
	suite.Equal("456", value.RemappedUser.ID)
	suite.Equal("Walter", value.RemappedUser.Name)

	suite.Equal(int8(0), value.Int8)
	suite.Equal(false, value.RemappedUser.AuditTrail.Deleted)
}

func (suite *JSONSuite) TestDecodeEncodeValues() {
	inTime := time.Date(2010, time.November, 11, 12, 0, 0, 0, time.UTC)
	inTimePtr := time.Date(2020, time.November, 11, 12, 0, 0, 0, time.UTC)

	values := codec.JSONEncoder{}.EncodeValues(testext.SampleComplexRequest{
		InFlag:  true,
		InFloat: 3.14,
		InUser: testext.SampleUser{
			ID:              "abc",
			Name:            "The Dude",
			Age:             47,
			Attention:       5 * time.Second,
			AttentionString: testext.CustomDuration(4 * time.Minute),
			PhoneNumber:     "555-1234",
			MarshalToString: testext.MarshalToString{
				Home: "home@string.com",
				Work: "work@string.com",
			},
			MarshalToObject: testext.MarshalToObject{
				Home: "home@object.com",
				Work: "work@object.com",
			},
		},
		InTime:    inTime,
		InTimePtr: &inTimePtr,
	})

	out := testext.SampleComplexRequest{}
	err := codec.JSONDecoder{}.DecodeValues(values, &out)
	suite.Require().NoError(err)
	suite.Equal(true, out.InFlag)
	suite.Equal(3.14, out.InFloat)
	suite.Equal("abc", out.InUser.ID)
	suite.Equal("The Dude", out.InUser.Name)
	suite.Equal(47, out.InUser.Age)
	suite.Equal(5*time.Second, out.InUser.Attention)
	suite.Equal(testext.CustomDuration(4*time.Minute), out.InUser.AttentionString)
	suite.Equal("555-1234", out.InUser.PhoneNumber)
	suite.Equal("home@string.com", out.InUser.MarshalToString.Home)
	suite.Equal("work@string.com", out.InUser.MarshalToString.Work)
	suite.Equal("home@object.com", out.InUser.MarshalToObject.Home)
	suite.Equal("work@object.com", out.InUser.MarshalToObject.Work)
	suite.Equal(inTime, out.InTime)
	suite.Require().NotNil(out.InTimePtr)
	suite.Equal(inTimePtr, *out.InTimePtr)
}

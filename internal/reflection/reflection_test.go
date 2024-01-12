package reflection_test

import (
	"reflect"
	"testing"

	"github.com/bridgekitio/frodo/internal/reflection"
	"github.com/stretchr/testify/suite"
)

func TestReflectionSuite(t *testing.T) {
	suite.Run(t, new(ReflectionSuite))
}

type ReflectionSuite struct {
	suite.Suite
}

func (suite *ReflectionSuite) TestToBindingValue() {
	r := suite.Require()

	type organization struct {
		ID   int
		Name string
	}

	type group struct {
		ID     int
		Name   string
		Org    organization `json:"Organization"`
		OrgPtr *organization
	}

	type EmbeddedStruct struct {
		Foo string
		Bar int
	}

	type EmbeddedString string
	type EmbeddedStringAlias string
	type EmbeddedInt int

	type user struct {
		ID    int
		Name  string `json:"alias"`
		Group group
		EmbeddedStruct
		EmbeddedString
		EmbeddedStringAlias `json:"embedded_alias"`
		EmbeddedInt
	}

	// empty := user{}
	dude := user{
		ID:   123,
		Name: "Dude",
		Group: group{
			ID:   456,
			Name: "Bowling League",
			Org: organization{
				ID:   789,
				Name: "Lebowski-Fest",
			},
			OrgPtr: &organization{
				ID:   42,
				Name: "His Dudeness",
			},
		},
		EmbeddedStruct: EmbeddedStruct{
			Foo: "Foo!",
			Bar: 1000,
		},
		EmbeddedString:      "Blah",
		EmbeddedStringAlias: "Farts",
		EmbeddedInt:         9999,
	}

	testValidBinding := func(u user, path string, expected any, outValue any) {
		r.True(reflection.ToBindingValue(u, path, outValue))
		// The outValue is likely a pointer (e.g. &myEmbeddedString), so dereference the pointer for the comparison.
		r.EqualValues(expected, reflect.ValueOf(outValue).Elem().Interface())
	}

	var intValue int
	var stringValue string
	var embeddedIntValue EmbeddedInt
	var embeddedStringValue EmbeddedString
	var embeddedStringAliasValue EmbeddedStringAlias

	// Garbage data tests
	r.False(reflection.ToBindingValue(dude, "Turds", &intValue))
	r.False(reflection.ToBindingValue(dude, "", &intValue))
	r.False(reflection.ToBindingValue(stringValue, "ID", &intValue))
	r.False(reflection.ToBindingValue(dude, "ID", &stringValue))
	r.Panics(func() {
		reflection.ToBindingValue(dude, "ID", nil)
	})

	// Can properly fetch primitive fields at the root level
	testValidBinding(dude, "ID", 123, &intValue)
	testValidBinding(dude, "alias", "Dude", &stringValue)

	// Can go recursively deep for values
	testValidBinding(dude, "Group.ID", 456, &intValue)
	testValidBinding(dude, "Group.Name", "Bowling League", &stringValue)
	testValidBinding(dude, "Group.Organization.ID", 789, &intValue)
	testValidBinding(dude, "Group.OrgPtr.ID", 42, &intValue)
	testValidBinding(dude, "Group.OrgPtr.Name", "His Dudeness", &stringValue)

	// Properly deals with embedded types. You can reference it using the shorthand (pretend the embedding isn't
	// there), or explicitly name the embedded type.
	testValidBinding(dude, "Foo", "Foo!", &stringValue)
	testValidBinding(dude, "EmbeddedStruct.Foo", "Foo!", &stringValue)
	testValidBinding(dude, "Bar", 1000, &intValue)

	// You should be able to embed non-struct types and have them resolve if you treat it like a non-embedded field.
	// Also, embedded fields should work even if you give them JSON aliases as well!
	testValidBinding(dude, "EmbeddedString", EmbeddedString("Blah"), &embeddedStringValue)
	testValidBinding(dude, "embedded_alias", EmbeddedStringAlias("Farts"), &embeddedStringAliasValue)
	testValidBinding(dude, "EmbeddedInt", EmbeddedInt(9999), &embeddedIntValue)

	// Can grab complex data structures as binding values.
	var groupValue group
	r.True(reflection.ToBindingValue(dude, "Group", &groupValue))
	r.Equal(456, groupValue.ID)
	r.Equal("Bowling League", groupValue.Name)

	var orgValue organization
	r.True(reflection.ToBindingValue(dude, "Group.Organization", &orgValue))
	r.Equal(789, orgValue.ID)
	r.Equal("Lebowski-Fest", orgValue.Name)

	var orgPtrValue *organization
	r.True(reflection.ToBindingValue(dude, "Group.OrgPtr", &orgPtrValue))
	r.Equal(42, orgPtrValue.ID)
	r.Equal("His Dudeness", orgPtrValue.Name)

	// When a field is remapped using the `json` tag, you need to use that in the binding path. The actual
	// field name should NOT work!
	r.False(reflection.ToBindingValue(dude, "Name", &stringValue))
	r.False(reflection.ToBindingValue(dude, "Group.Org", &intValue))
	r.False(reflection.ToBindingValue(dude, "Group.Org.ID", &intValue))
}

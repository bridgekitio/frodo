package reflection

import (
	"reflect"
	"strings"
)

// noField is the default value returned by FindField when no field names match.
var noField = reflect.StructField{}

// IsStructOrPointerTo returns true if the given type is any sort of struct or a pointer
// to some sort of struct type.
func IsStructOrPointerTo(valueType reflect.Type) bool {
	if valueType.Kind() == reflect.Struct {
		return true
	}
	return valueType.Kind() == reflect.Ptr && valueType.Elem().Kind() == reflect.Struct
}

// IndirectTypeKind returns the underlying reflection Kind. This will unravel pointers
// until we get to a primordial value type.
//
//	IndirectTypeKind(reflect.Type(someInt))             -> reflect.Int
//	IndirectTypeKind(reflect.Type(someIntPointer))      -> reflect.Int
//	IndirectTypeKind(reflect.Type(someString))          -> reflect.String
//	IndirectTypeKind(reflect.Type(someStringPointer))   -> reflect.String
//	IndirectTypeKind(reflect.Type(someDuration))        -> reflect.Int64
//	IndirectTypeKind(reflect.Type(someDurationPointer)) -> reflect.Int64
func IndirectTypeKind(valueType reflect.Type) reflect.Kind {
	if valueType.Kind() == reflect.Ptr {
		return IndirectTypeKind(valueType.Elem())
	}
	return valueType.Kind()
}

// IsNil returns true if the given value's type is both nil-able and nil.
func IsNil(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

// FindField looks up the struct field attribute for the given field on the given struct.
func FindField(structType reflect.Type, name string) (reflect.StructField, bool) {
	if structType.Kind() != reflect.Struct {
		return noField, false
	}
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if strings.EqualFold(name, BindingName(field)) {
			return field, true
		}
		if !field.Anonymous {
			continue
		}
		if embeddedField, ok := FindField(field.Type, name); ok {
			return embeddedField, ok
		}
	}
	return noField, false
}

// BindingName just returns the name of the field/attribute on the struct unless it has a `json` tag
// defined. If so, it will use the remapped name for this field instead.
//
//	type Foo struct {
//	    A string
//	    B string `json:"hello"
//	}
//
// The binding name for the first attribute is "A", but the binding name for the other is "hello".
func BindingName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return field.Name
	}

	// Parse the `json` tag to determine how the user has re-mapped the field.
	switch comma := strings.IndexRune(jsonTag, ','); comma {
	case -1:
		// e.g. `json:"firstName"`
		return jsonTag
	case 0:
		// e.g. `json:",omitempty"` (not remapped so use fields actual name)
		return field.Name
	default:
		// e.g. `json:"firstName,omitempty" (just use the remapped name)
		return jsonTag[0:comma]
	}
}

// FlattenPointerType looks at the reflective type and if it's a pointer it will flatten it to the
// type it is a pointer for (e.g. "*string"->"string"). If it's already a non-pointer then we will
// leave this type as-is.
func FlattenPointerType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

// Assign simply performs a reflective replacement of the value, making sure to try to properly handle pointers.
func Assign(value any, out any) bool {
	// Depending on whether you wrote "SomeStruct{}" or "&SomeStruct{}" (a pointer) to the
	// scope, we want to make sure that we're de-referencing the scope value properly.
	reflectValue := reflect.ValueOf(value)
	reflectOut := reflect.ValueOf(out)

	valueIsPointer := reflectValue.Type().Kind() == reflect.Ptr
	outIsPointerToPointer := reflectOut.Type().Kind() == reflect.Ptr

	switch {
	case valueIsPointer && outIsPointerToPointer:
		return set(reflectValue, reflectOut.Elem())
	case valueIsPointer:
		return set(reflectValue.Elem(), reflectOut.Elem())
	default:
		return set(reflectValue, reflectOut.Elem())
	}
}

func set(value reflect.Value, out reflect.Value) bool {
	// Should have no problem assigning "foo" to an out variable of type &string
	if out.Type().AssignableTo(value.Type()) {
		out.Set(value)
		return true
	}
	// If your value is a type alias of the out type, convert and set. For instance:
	// your value type is 'string' and your out is the alias 'type ID string'. Convert your
	// string into an ID and then set the value.
	if out.Type().ConvertibleTo(value.Type()) {
		out.Set(value.Convert(out.Type()))
		return true
	}
	return false
}

// ToBindingValue is used for fetching one-off values from struct instances given their binding string. For
// example, you can take an instance of a User{} struct and the binding path "Group.ID" to fetch the user's group id.
//
// Example:
//
//	user := User{Name: "Bob", Group: Group{ID:"12345", Name:"Admins"}}
//	name := ToBindingValue(user, "Name") // will be "Bob"
//	groupID := ToBindingValue(user, "Group.ID") // will be "12345"
//	groupName := ToBindingValue(user, "Group.Name") // will be "Admins"
func ToBindingValue(value any, bindingPath string, out any) bool {
	if bindingPath == "" {
		return Assign(value, out)
	}

	nextAttribute, remainingPath, _ := strings.Cut(bindingPath, ".")
	reflectValue := reflect.ValueOf(value)
	reflectType := reflectValue.Type()

	// Flatten pointers to the raw value.
	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
		reflectValue = reflectValue.Elem()
	}

	bindingValue, ok := resolveBindingValue(reflectValue, reflectType, nextAttribute)
	if !ok {
		return false
	}
	return ToBindingValue(bindingValue, remainingPath, out)
}

func isStructType(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func resolveBindingValue(structValue reflect.Value, structType reflect.Type, name string) (any, bool) {
	if !isStructType(structType) {
		return nil, false
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		bindingName := BindingName(field)

		// There's a field on this struct whose name or JSON tag name matches the binding field name. Got it!
		if strings.EqualFold(name, bindingName) {
			return structValue.Field(i).Interface(), true
		}

		// The other possibility is that the field we're iterating over is an embedded field, so allow fields on
		// the embedded type to act as though they are named fields on this struct. For example:
		//
		// type User struct {
		//     Grouped
		//     Enabled
		//     ID string
		//     Name string
		// }
		//
		// type Grouped struct {
		//     GroupIDs []string
		// }
		//
		// type Enabled bool
		//
		// If we are doing resolveBindingValue(someUser, userType, "GroupIDs"), we want to recursively try to
		// resolve the fields on the Grouped type since it's embedded in User (i.e. "GroupIDs" should behave like
		// it's a native field on User).
		if !field.Anonymous {
			continue
		}

		// Only do recursive lookups on struct types. In the example above, we can't recursively iterate on
		// the Enabled type because even though it's embedded, it's not a struct. The only way to match that
		// field is to resolve the binding value on the "Enabled" field. It's possible to bind embedded non-structs
		// like the Enabled field, but the only way to catch those cases is above when we compare the binding name
		// to the field name.
		if !isStructType(field.Type) {
			continue
		}
		if embeddedValue, ok := resolveBindingValue(structValue.Field(i), field.Type, name); ok {
			return embeddedValue, ok
		}
	}
	return nil, false
}

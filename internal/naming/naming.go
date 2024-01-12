package naming

import (
	"strings"
	"unicode"
)

// NoPackage strips of any package prefixes from an identifier (e.g. "context.Context" -> "Context")
func NoPackage(ident string) string {
	period := strings.LastIndex(ident, ".")
	if period < 0 {
		return ident
	}
	return ident[period+1:]
}

// NoPointer strips off any "*" prefix your type identifier might have (e.g. "*Foo" -> "Foo")
func NoPointer(ident string) string {
	return strings.TrimLeft(ident, "*")
}

// NoSlice takes a string like "[]Foo" or "[456]Foo", strips off the slice/array braces, leaving you with "Foo".
func NoSlice(ident string) string {
	closeBrace := strings.Index(ident, "]")
	if closeBrace < 0 {
		return ident
	}
	return ident[closeBrace+1:]
}

// JoinPackageName converts a package-qualified type such as "fmt.Stringer" into a single "safe" identifier
// such as "fmtStringer". This is useful when converting types to languages with different naming semantics.
func JoinPackageName(ident string) string {
	return strings.ReplaceAll(ident, ".", "")
}

// NoImport strips off the prefix before "/" in your type identifier (e.g. "github.com/foo/bar/Baz" -> "Baz")
func NoImport(ident string) string {
	if slash := strings.LastIndex(ident, "/"); slash >= 0 {
		ident = ident[slash+1:]
	}
	return ident
}

// CleanPrefix strips the "command-line-arguments." prefix that the Go 'packages' package prepends to type
// identifier for types defined in the source file we're parsing.
func CleanPrefix(ident string) string {
	ident = strings.ReplaceAll(ident, "command-line-arguments.", "")
	ident = strings.ReplaceAll(ident, "command-line-arguments", "")
	return ident
}

// LeadingSlash adds... a leading slash to the given string.
func LeadingSlash(value string) string {
	if strings.HasPrefix(value, "/") {
		return value
	}
	return "/" + value
}

// ToLowerCamel converts the string to lower camel-cased.
func ToLowerCamel(value string) string {
	// This is a shitty implementation.
	if value == "" {
		return ""
	}

	lastUpperIndex := -1
	for i, r := range value {
		if !unicode.IsUpper(r) {
			// as soon as we come across a non-upper case rune we don't need to look anymore
			break
		}
		lastUpperIndex = i
	}

	switch {
	case lastUpperIndex < 0:
		// It's already all lower case...
		return value
	case lastUpperIndex == 0:
		// Looks like we're changing "FooBar" into "fooBar"
		return strings.ToLower(value[0:1]) + value[1:]
	case lastUpperIndex == len(value)-1:
		// Looks like we're changing the whole thing from upper to lower (e.g. "FOO" to "foo", or "ID" to "id")
		return strings.ToLower(value)
	default:
		// Looks like we're changing "HTTPClient" into "httpClient"
		return strings.ToLower(value[0:lastUpperIndex]) + value[lastUpperIndex:]
	}
}

// ToUpperCamel converts the string to upper camel-cased.
func ToUpperCamel(value string) string {
	// This is a shitty implementation.
	if value == "" {
		return ""
	}

	firstChar := value[0:1]
	return strings.ToUpper(firstChar) + value[1:]
}

// EmptyString is a predicate that returns true when the input value is "".
func EmptyString(value string) bool {
	return value == ""
}

// NotEmptyString is a predicate that returns true when the input value is anything but "".
func NotEmptyString(value string) bool {
	return value != ""
}

// PathTokens accepts a path string like "foo/bar/baz" and returns a slice of the individual
// path segment tokens such as ["foo", "bar", "baz"]. This will ignore leading/trailing slashes
// in your path so that you don't get leading/trailing "" tokens in your slice. This does not,
// however, clean up empty tokens caused by "//" somewhere in your path.
func PathTokens(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

// CleanTypeNameUpper normalizes a raw type's name to be a single token name in upper camel case.
func CleanTypeNameUpper(typeName string) string {
	typeName = CleanPrefix(typeName)
	typeName = NoSlice(typeName)
	typeName = NoPointer(typeName)
	typeName = JoinPackageName(typeName)

	// A bad-but-cheap "ToUpper" that only worries about the very first rune.
	firstChar := typeName[0:1]
	return strings.ToUpper(firstChar) + typeName[1:]
}

// DispositionFileName extracts the "filename" from an HTTP Content-Disposition header value.
func DispositionFileName(contentDisposition string) string {
	// The start or the file name in the header is the index of "filename=" plus the 9
	// characters in that substring.
	fileNameAttrIndex := strings.Index(contentDisposition, "filename=")
	if fileNameAttrIndex < 0 {
		return ""
	}

	// Support the fact that all of these are valid for the disposition header:
	//
	//   attachment; filename=foo.pdf
	//   attachment; filename="foo.pdf"
	//   attachment; filename='foo.pdf'
	//
	// This just makes sure that you don't have any quotes in your final value.
	fileName := contentDisposition[fileNameAttrIndex+9:]
	fileName = strings.Trim(fileName, `"'`)
	fileName = strings.ReplaceAll(fileName, `\"`, `"`)
	return fileName
}

// CleanFileName accepts a proposed file name like "Foo @ Bar.pdf", and normalizes it to a safe, valid file
// name like "Foo__Bar.pdf". The rules for normalization are:
//
// - Alphanumeric characters are okay
// - Underscores are okay
// - Periods are okay
// - Hyphens are okay
// - Whitespace is converted to underscores
func CleanFileName(fileName string) string {
	builder := strings.Builder{}
	for _, r := range fileName {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r), r == '_', r == '.', r == '-':
			builder.WriteRune(r)
		case unicode.IsSpace(r):
			builder.WriteRune('_')
		default:
			// Don't write potentially problematic characters. There should be plenty
			// of alphanumeric ones to make the file name seem reasonable enough.
		}
	}
	return builder.String()
}

// IsPathVariable returns true if the given path segment value looks like it's a variable.
//
//	IsPathVariable("user") -> false
//	IsPathVariable(":user") -> false
//	IsPathVariable("{ID}") -> true
//	IsPathVariable("{Group.Organization.ID}") -> true
func IsPathVariable(token string) bool {
	// The length check ensures we don't allow "{}" since there's no variable in there.
	return len(token) > 2 && strings.HasPrefix(token, "{") && strings.HasSuffix(token, "}")
}

// PathVariableName returns the raw variable name of the path variable if it actually is a variable (e.g. "{Foo.ID"}"
// becomes "Foo.ID). If it's not a path variable, then this return "" (e.g "User" -> "").
func PathVariableName(pathSegment string) string {
	if IsPathVariable(pathSegment) {
		return pathSegment[1 : len(pathSegment)-1]
	}
	return ""
}

// ResolvePath accepts a path that potentially includes path variables (e.g. "{ID}) and uses the 'variableFunc' to
// replace those tokens in the path w/ some runtime value.
//
//	ResolvePath("foo/BAR", '/', strings.ToLower) --> "foo/BAR")
//	ResolvePath("foo/{BAR}", '/', strings.ToLower) --> "foo/bar")
func ResolvePath(path string, delim rune, variableFunc func(string) string) string {
	segments := TokenizePath(path, '.')
	for i, segment := range segments {
		if pathVar := PathVariableName(segment); pathVar != "" {
			segments[i] = variableFunc(pathVar)
		}
	}
	return strings.Join(segments, string(delim))
}

// TokenizePath follows our standard segment/variable naming conventions and splits a path such as "foo/{bar/baz}/goo"
// or "foo.{bar.baz}.goo" into the slice ["foo", "{bar.baz}", "goo"].
func TokenizePath(path string, delim rune) []string {
	// Allocate a slice w/ 1 more than the number of delimiters. This means that for a path like "foo/bar/baz",
	// there are 2 slashes; we'll allocate a slice w/ capacity 3. That's perfect for most cases, but you might
	// do a little over-allocation when parsing roles like "group.{User.Group.ID}.write" which would allocate 5
	// elements even though in the end we'll only need 3. That's okay.
	results := make([]string, 0, strings.Count(path, string(delim))+1)

	// We track whether we're parsing a variable or not so in instances like "foo.{bar.baz}.goo", we know to
	// ignore the "." in "{bar.baz}" and treat that whole thing like one token.
	variableDepth := 0
	currentToken := strings.Builder{}

	for _, char := range path {
		switch {
		// We're finished parsing a path variable. Aside from closing out the token, it means that we will
		// stop ignoring instances of the delim and treat them like the separators they are again.
		case variableDepth > 0 && char == '}':
			currentToken.WriteRune(char)
			variableDepth--

		// For some dumb-shit reason you have a path like "foo/{bar{baz}goo}/blah". This notices that until you
		// encounter the final closing '}', you're still technically parsing the variable. Increase the depth so
		// that when we hit the first closing brace, we're still parsing.
		case variableDepth > 0 && char == '{':
			currentToken.WriteRune(char)
			variableDepth++

		case variableDepth > 0:
			currentToken.WriteRune(char)

		case char == '{':
			currentToken.WriteRune(char)
			variableDepth++

		case char == delim:
			results = append(results, currentToken.String())
			currentToken.Reset()

		default:
			currentToken.WriteRune(char)
		}
	}

	// Make sure to close out the last token that we were in the middle of ingesting.
	if currentToken.Len() == 0 {
		return results
	}
	return append(results, currentToken.String())
}

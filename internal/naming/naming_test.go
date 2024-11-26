//go:build unit

package naming_test

import (
	"testing"

	"github.com/bridgekit-io/frodo/internal/naming"
	"github.com/stretchr/testify/suite"
)

func TestNamingSuite(t *testing.T) {
	suite.Run(t, new(NamingSuite))
}

type NamingSuite struct {
	suite.Suite
}

func (suite *NamingSuite) TestNoPackage() {
	r := suite.Require()
	r.Equal("", naming.NoPackage(""))
	r.Equal("foo", naming.NoPackage("foo"))
	r.Equal("foo_bar", naming.NoPackage("foo_bar"))
	r.Equal("bar", naming.NoPackage("foo.bar"))
	r.Equal("baz", naming.NoPackage("foo.bar.baz"))
	r.Equal("baz", naming.NoPackage("foo.bar...baz"))
	r.Equal("baz", naming.NoPackage("*foo.bar...baz"))
}

func (suite *NamingSuite) TestNoPointer() {
	r := suite.Require()
	r.Equal("", naming.NoPointer(""))
	r.Equal("foo", naming.NoPointer("foo"))
	r.Equal("foo_bar", naming.NoPointer("foo_bar"))
	r.Equal("foo*bar", naming.NoPointer("foo*bar")) // only strip from left
	r.Equal("foo*", naming.NoPointer("foo*"))
	r.Equal("foo", naming.NoPointer("*foo"))
	r.Equal("foo", naming.NoPointer("**foo"))
	r.Equal(" *foo", naming.NoPointer("* *foo")) // only works on single tokens
	r.Equal("foo.bar.baz", naming.NoPointer("**foo.bar.baz"))
	r.Equal("&foo", naming.NoPointer("&foo")) // only strip pointer declarations, not references
	r.Equal("&foo", naming.NoPointer("*&foo"))
}

func (suite *NamingSuite) TestJoinPackageName() {
	r := suite.Require()
	r.Equal("", naming.JoinPackageName(""))
	r.Equal("foo", naming.JoinPackageName("foo"))
	r.Equal("foo bar", naming.JoinPackageName("foo bar"))
	r.Equal("foobar", naming.JoinPackageName("foo.bar"))
	r.Equal("foobarbaz", naming.JoinPackageName("foo.bar.baz"))
	r.Equal("fooBar", naming.JoinPackageName("foo.Bar"))
	r.Equal("fooBar", naming.JoinPackageName("foo..Bar"))
	r.Equal("*fooBar", naming.JoinPackageName("*foo.Bar")) // you have to do your own "un-pointer-ing"
}

func (suite *NamingSuite) TestNoImport() {
	r := suite.Require()
	r.Equal("", naming.NoImport(""))
	r.Equal("foo", naming.NoImport("foo"))
	r.Equal("foo.Bar", naming.NoImport("foo.Bar"))
	r.Equal("baz", naming.NoImport("foo/bar/baz"))
	r.Equal("", naming.NoImport("foo/bar/baz/")) // assume trailing slash means missing identifier name
	r.Equal("baz", naming.NoImport("foo/  /  /bar///baz"))
	r.Equal("baz.Blah", naming.NoImport("foo/bar/baz.Blah"))
}

func (suite *NamingSuite) TestCleanPrefix() {
	r := suite.Require()
	r.Equal("", naming.CleanPrefix(""))
	r.Equal("foo", naming.CleanPrefix("foo"))
	r.Equal("foo.bar", naming.CleanPrefix("foo.bar"))
	r.Equal("*foo.Bar", naming.CleanPrefix("*foo.Bar"))
	r.Equal("*foo/bar.Baz", naming.CleanPrefix("*foo/bar.Baz"))
	r.Equal("*foo/bar.Baz", naming.CleanPrefix("*foo/bar.Baz"))
	r.Equal("", naming.CleanPrefix("command-line-arguments"))
	r.Equal("", naming.CleanPrefix("command-line-arguments."))
	r.Equal("*", naming.CleanPrefix("*command-line-arguments"))
	r.Equal("****", naming.CleanPrefix("****command-line-arguments."))
	r.Equal("foo.Bar", naming.CleanPrefix("command-line-arguments.foo.Bar"))
	r.Equal("*foo.Bar", naming.CleanPrefix("*command-line-arguments.foo.Bar"))
	r.Equal("****foo.Bar", naming.CleanPrefix("****command-line-arguments.foo.Bar"))
	r.Equal("COMMAND-line-arguments.foo.Bar", naming.CleanPrefix("COMMAND-line-arguments.foo.Bar")) // case sensitive - not how go gives it to us
	r.Equal("[]foo.Bar", naming.CleanPrefix("[]command-line-arguments.foo.Bar"))
	r.Equal("[]*foo.Bar", naming.CleanPrefix("[]*command-line-arguments.foo.Bar"))
	r.Equal("[123]*foo.Bar", naming.CleanPrefix("[123]*command-line-arguments.foo.Bar"))
}

func (suite *NamingSuite) TestLeadingSlash() {
	r := suite.Require()
	r.Equal("/", naming.LeadingSlash(""))
	r.Equal("/", naming.LeadingSlash("/"))
	r.Equal("//////", naming.LeadingSlash("//////"))
	r.Equal("/foo", naming.LeadingSlash("foo"))
	r.Equal("/foo.bar", naming.LeadingSlash("foo.bar"))
	r.Equal("/foo/bar", naming.LeadingSlash("foo/bar"))
	r.Equal("/foo/bar//", naming.LeadingSlash("/foo/bar//"))
	r.Equal("//foo/bar//", naming.LeadingSlash("//foo/bar//"))
}

func (suite *NamingSuite) TestEmptyString() {
	r := suite.Require()
	r.Equal(true, naming.EmptyString(""))
	r.Equal(false, naming.EmptyString(" "))
	r.Equal(false, naming.EmptyString("/"))
	r.Equal(false, naming.EmptyString("foo"))
	r.Equal(false, naming.EmptyString("🍺"))
}

func (suite *NamingSuite) TestNotEmptyString() {
	r := suite.Require()
	r.Equal(false, naming.NotEmptyString(""))
	r.Equal(true, naming.NotEmptyString(" "))
	r.Equal(true, naming.NotEmptyString("/"))
	r.Equal(true, naming.NotEmptyString("foo"))
	r.Equal(true, naming.NotEmptyString("🍺"))
}

func (suite *NamingSuite) TestPathTokens() {
	r := suite.Require()
	r.Equal([]string{}, naming.PathTokens(""))
	r.Equal([]string{}, naming.PathTokens("/"))
	r.Equal([]string{"*"}, naming.PathTokens("*"))
	r.Equal([]string{"foo"}, naming.PathTokens("foo"))
	r.Equal([]string{"foo"}, naming.PathTokens("/foo"))
	r.Equal([]string{"foo"}, naming.PathTokens("/foo/"))
	r.Equal([]string{"foo"}, naming.PathTokens("foo/"))
	r.Equal([]string{"foo", "bar"}, naming.PathTokens("foo/bar"))
	r.Equal([]string{"foo", "{bar}"}, naming.PathTokens("foo/{bar}"))
	r.Equal([]string{"foo", "{bar}", "baz"}, naming.PathTokens("foo/{bar}/baz"))
	r.Equal([]string{"foo", "{bar}", "", "", "baz"}, naming.PathTokens("foo/{bar}///baz")) // doesn't normalize inner /
	r.Equal([]string{"foo", "{bar}", "{baz.Blah}"}, naming.PathTokens("foo/{bar}/{baz.Blah}"))
	r.Equal([]string{"foo", "{bar}", "{baz.Blah}"}, naming.PathTokens("///foo/{bar}/{baz.Blah}///"))
}

func (suite *NamingSuite) TestLowerCamel() {
	r := suite.Require()
	r.Equal("", naming.ToLowerCamel(""))
	r.Equal("foo", naming.ToLowerCamel("foo"))
	r.Equal("foo", naming.ToLowerCamel("Foo"))
	r.Equal("id", naming.ToLowerCamel("ID"))
	r.Equal("fooBar", naming.ToLowerCamel("fooBar"))
	r.Equal("fooBAR", naming.ToLowerCamel("fooBAR"))
	r.Equal("fooBAR", naming.ToLowerCamel("FooBAR"))
	r.Equal("fOoBAR", naming.ToLowerCamel("FOoBAR"))
	r.Equal("httpClient", naming.ToLowerCamel("HTTPClient"))
	r.Equal("5OOBAR", naming.ToLowerCamel("5OOBAR"))
}

func (suite *NamingSuite) TestUpperCamel() {
	r := suite.Require()
	r.Equal("", naming.ToUpperCamel(""))
	r.Equal("Foo", naming.ToUpperCamel("foo"))
	r.Equal("Foo", naming.ToUpperCamel("Foo"))
	r.Equal("Id", naming.ToUpperCamel("id"))
	r.Equal("FooBar", naming.ToUpperCamel("fooBar"))
	r.Equal("FooBAR", naming.ToUpperCamel("fooBAR"))
	r.Equal("FooBAR", naming.ToUpperCamel("FooBAR"))
	r.Equal("5OOBAR", naming.ToUpperCamel("5OOBAR"))
}

func (suite *NamingSuite) TestIsPathVariable() {
	r := suite.Require()
	r.False(naming.IsPathVariable(""))
	r.False(naming.IsPathVariable("foo"))
	r.False(naming.IsPathVariable(":foo"))
	r.False(naming.IsPathVariable("{foo"))  // no closing brace
	r.False(naming.IsPathVariable("foo}"))  // no opening brace
	r.False(naming.IsPathVariable("f{o}o")) // braces must be at the start/end of the string
	r.False(naming.IsPathVariable("{}"))    // must have a variable name inside the braces

	r.True(naming.IsPathVariable("{foo}"))
	r.True(naming.IsPathVariable("{foo.bar}"))
	r.True(naming.IsPathVariable("{Foo.Bar.Baz}"))
	r.True(naming.IsPathVariable("{Foo.{Bar}.Baz}")) // technically true but will break at runtime most likely
}

func (suite *NamingSuite) TestPathVariableName() {
	r := suite.Require()
	r.Equal("", naming.PathVariableName(""))
	r.Equal("", naming.PathVariableName("foo"))
	r.Equal("", naming.PathVariableName(":foo"))
	r.Equal("", naming.PathVariableName("{foo"))
	r.Equal("", naming.PathVariableName("foo}"))
	r.Equal("", naming.PathVariableName("f{o}o"))
	r.Equal("", naming.PathVariableName("{}"))

	r.Equal("foo", naming.PathVariableName("{foo}"))
	r.Equal("foo.bar", naming.PathVariableName("{foo.bar}"))
	r.Equal("Foo.Bar.Baz", naming.PathVariableName("{Foo.Bar.Baz}"))
	r.Equal("Foo.{Bar}.Baz", naming.PathVariableName("{Foo.{Bar}.Baz}"))
}

func (suite *NamingSuite) TestTokenizePath() {
	testCase := func(path string, delim rune, expectedTokens []string) {
		suite.Require().Equal(expectedTokens, naming.TokenizePath(path, delim))
	}

	testCase("", '/', []string{})

	// Paths without variables
	testCase("foo", '/', []string{"foo"})
	testCase("foo/bar", '/', []string{"foo", "bar"})
	testCase("foo/bar/baz.goo", '/', []string{"foo", "bar", "baz.goo"})
	testCase("foo.bar.baz.goo", '/', []string{"foo.bar.baz.goo"})
	testCase("foo.bar.baz.goo", '.', []string{"foo", "bar", "baz", "goo"})

	// Paths with variables
	testCase("{foo}", '/', []string{"{foo}"})
	testCase("{foo}/bar/{baz}", '/', []string{"{foo}", "bar", "{baz}"})
	testCase("{foo/moo}/bar/{baz}", '/', []string{"{foo/moo}", "bar", "{baz}"})
	testCase("{foo.moo}.bar.{baz.raz}", '.', []string{"{foo.moo}", "bar", "{baz.raz}"})

	// How to treat trailing delimiters
	testCase("/", '/', []string{""})
	testCase("foo///", '/', []string{"foo", "", ""})

	// Weirdo cases
	testCase("foo///junk/bar/{baz}", '/', []string{"foo", "", "", "junk", "bar", "{baz}"})
	testCase("{foo}junk/bar/{baz}", '/', []string{"{foo}junk", "bar", "{baz}"})
	testCase("{f{o}o}junk/bar/{baz}", '/', []string{"{f{o}o}junk", "bar", "{baz}"})
	testCase("{f{o}/o}junk/bar/{baz}", '/', []string{"{f{o}/o}junk", "bar", "{baz}"})
	testCase("{f{o}/o}junk/bar/{baz}", '.', []string{"{f{o}/o}junk/bar/{baz}"})
}

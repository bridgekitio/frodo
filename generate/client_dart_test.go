//go:build integration

package generate_test

import (
	"testing"
	"time"

	"github.com/bridgekitio/frodo/internal/testext"
	"github.com/stretchr/testify/suite"
)

func TestDartClientSuite(t *testing.T) {
	suite.Run(t, &DartClientSuite{GeneratedClientSuite{
		addresses: testext.NewFreeAddress("localhost", 20100),
	}})
}

type DartClientSuite struct {
	GeneratedClientSuite
}

func (suite *DartClientSuite) Run(address string, testName string, expectedLines int) ClientTestResults {
	output := suite.RunExternalTest("dart testdata/dart/run_client.dart " + testName + " " + address)
	suite.Require().Len(output, expectedLines, "Test runner output has wrong number of lines.")
	return output
}

// Ensures that we get a connection refused error when connecting to a not-running server.
func (suite *DartClientSuite) TestNotConnected() {
	output := suite.Run("NotConnected", "localhost:20100", 1)
	suite.ExpectFail(output[0], 500, "connection refused")
}

// Ensures that we can rely on all default behaviors for an endpoint.
func (suite *DartClientSuite) TestDefaults() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("Defaults", address, 1)
	res := testext.SampleResponse{}
	suite.ExpectPass(output[0], &res, func() {
		suite.Equal("Defaults:Dude", res.Text)
	})
}

// Ensures that we can encode/decode non-flat structs w/ nothing but strings.
func (suite *DartClientSuite) TestComplexValues() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("ComplexValues", address, 1)
	res := testext.SampleComplexResponse{}
	suite.ExpectPass(output[0], &res, func() {
		suite.Equal(true, res.OutFlag)
		suite.Equal(3.14, res.OutFloat)
		suite.Equal("123", res.OutUser.ID)
		suite.Equal("Dude", res.OutUser.Name)
		suite.Equal(47, res.OutUser.Age)
		suite.Equal(time.Duration(1000000), res.OutUser.Attention)
		suite.Equal("555-1234", res.OutUser.PhoneNumber)
		suite.Equal("home@string.com", res.OutUser.MarshalToString.Home)
		suite.Equal("work@string.com", res.OutUser.MarshalToString.Work)

		// These work in POST/PUT/PATCH calls. I'd prefer if people not use this
		// style either way, but for now it works.
		suite.Equal("home@object.com", res.OutUser.MarshalToObject.Home)
		suite.Equal("work@object.com", res.OutUser.MarshalToObject.Work)
	})
}

// Ensures that we can encode/decode non-flat structs w/ nothing but strings.
func (suite *DartClientSuite) TestComplexValuesPath() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("ComplexValuesPath", address, 1)
	res := testext.SampleComplexResponse{}
	suite.ExpectPass(output[0], &res, func() {
		suite.Equal(true, res.OutFlag)
		suite.Equal(3.14, res.OutFloat)
		suite.Equal("123", res.OutUser.ID)
		suite.Equal("Dude", res.OutUser.Name)
		suite.Equal(47, res.OutUser.Age)
		suite.Equal(time.Duration(1000000), res.OutUser.Attention)
		suite.Equal(testext.CustomDuration(4*time.Minute+2*time.Second), res.OutUser.AttentionString)
		suite.Equal("555-1234", res.OutUser.PhoneNumber)
		suite.Require().NotNil(res.OutTimePtr)
		suite.Equal("2020-11-06T17:47:12Z", res.OutTimePtr.Format(time.RFC3339))
		suite.Equal("2022-12-05T17:47:12Z", res.OutTime.Format(time.RFC3339))
		suite.Equal("home@string.com", res.OutUser.MarshalToString.Home)
		suite.Equal("work@string.com", res.OutUser.MarshalToString.Work)

		// KNOWN LIMITATION CHECK: The custom MarshalJSON/UnmarshalJSON implementation changes
		// the expected format, but Dart client has absolutely no way of knowing that. It just thinks
		// that the struct looks normal and will submit "InUser.MarshalToObject.H" even though
		// there's no way for our JSON value decoder to know that "H" maps to the Home field.
		suite.Equal("", res.OutUser.MarshalToObject.Home)
		suite.Equal("", res.OutUser.MarshalToObject.Work)
	})
}

// Ensures that the client reports back 4XX style errors when they're returned.
func (suite *DartClientSuite) TestFail4XX() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("Fail4XX", address, 1)
	suite.ExpectFail(output[0], 409, "always a conflict")
}

// Ensures that the client reports back 5XX style errors when they're returned.
func (suite *DartClientSuite) TestFail5XX() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("Fail5XX", address, 1)
	suite.ExpectFail(output[0], 502, "always a bad gateway")
}

// Ensures that we can define a custom method/path and still send data properly.
func (suite *DartClientSuite) TestCustomRoute() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("CustomRoute", address, 1)
	res := testext.SampleResponse{}
	suite.ExpectPass(output[0], &res, func() {
		suite.Equal("123", res.ID)
		suite.Equal("Route:Abide", res.Text)
	})
}

// Ensures that we can define a custom method/path and still send data properly.
func (suite *DartClientSuite) TestCustomRouteBody() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("CustomRouteBody", address, 1)
	res := testext.SampleResponse{}
	suite.ExpectPass(output[0], &res, func() {
		suite.Equal("123", res.ID)
		suite.Equal("Route:Abide", res.Text)
	})
}

// Ensures that we can define a custom method/path and still send data properly.
func (suite *DartClientSuite) TestCustomRouteQuery() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("CustomRouteQuery", address, 1)
	res := testext.SampleResponse{}
	suite.ExpectPass(output[0], &res, func() {
		suite.Equal("123", res.ID)
		suite.Equal("Route:Abide", res.Text)
	})
}

// Ensures that the client fails if you attempt to invoke a function that has "HTTP OMIT" on it.
func (suite *DartClientSuite) TestOmitMe() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("OmitMe", address, 1)
	suite.ExpectFail(output[0], 501, "")
}

// Ensures that the client can handle receiving a raw stream of data rather than auto-encoding.
func (suite *DartClientSuite) TestDownload() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("Download", address, 1)
	res := RawClientOutput{}
	suite.ExpectPass(output[0], &res, func() {
		suite.Equal("text/plain", res.ContentType)
		suite.Equal("dude.txt", res.ContentFileName)
		suite.Equal(34, res.ContentLength)
		suite.Equal("Donny, you're out of your element!", res.Content)
	})
}

// Ensures that the client can handle receiving a raw stream of data that includes range
// information, so you can resume the stream later.
func (suite *DartClientSuite) TestDownloadResumable() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("DownloadResumable", address, 1)
	res := RawClientOutput{}
	suite.ExpectPass(output[0], &res, func() {
		suite.Equal("text/html", res.ContentType)
		suite.Equal("dude.html", res.ContentFileName)
		suite.Equal(24, res.ContentLength)
		suite.Equal("<h1>The Dude Abides</h1>", res.Content)
		suite.Equal("bytes", res.ContentRange.Unit)
		suite.Equal(50, res.ContentRange.Start)
		suite.Equal(74, res.ContentRange.End)
		suite.Equal(1024, res.ContentRange.Size)
	})
}

// Ensures that the client can handle receiving a redirect and reads it in as a raw stream.
func (suite *DartClientSuite) TestRedirect() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("Redirect", address, 1)
	res := RawClientOutput{}
	suite.ExpectPass(output[0], &res, func() {
		suite.Equal("text/csv", res.ContentType)
		suite.Equal(42, res.ContentLength)
		suite.Equal("ID,Name,Enabled\n1,Dude,true\n2,Walter,false", res.Content)
	})
}

// Ensures that the client passes along authorization info specified when invoking a function.
func (suite *DartClientSuite) TestAuthorization() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("Authorization", address, 1)
	res := testext.SampleResponse{}
	suite.ExpectPass(output[0], &res, func() {
		suite.Equal("Abide", res.Text)
	})
}

// Ensures that the client passes along authorization info specified only during client creation.
func (suite *DartClientSuite) TestAuthorizationGlobal() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("AuthorizationGlobal", address, 1)
	res := testext.SampleResponse{}
	suite.ExpectPass(output[0], &res, func() {
		suite.Equal("12345", res.Text)
	})
}

// Ensures that the client passes along authorization info included during method invocation if
// it's supplied there AND when creating the client.
func (suite *DartClientSuite) TestAuthorizationOverride() {
	address, shutdown := suite.startServer()
	defer shutdown()

	output := suite.Run("AuthorizationOverride", address, 1)
	res := testext.SampleResponse{}
	suite.ExpectPass(output[0], &res, func() {
		suite.Equal("Abide", res.Text)
	})
}

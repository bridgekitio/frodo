//go:build integration

package generate_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bridgekitio/frodo/fail"
	"github.com/bridgekitio/frodo/internal/testext"
	gen "github.com/bridgekitio/frodo/internal/testext/gen"
	"github.com/bridgekitio/frodo/metadata"
	"github.com/bridgekitio/frodo/services/clients"
	"github.com/stretchr/testify/suite"
)

func TestGoClientSuite(t *testing.T) {
	suite.Run(t, &GoClientSuite{GeneratedClientSuite{
		addresses: testext.NewFreeAddress("localhost", 20300),
	}})
}

type GoClientSuite struct {
	GeneratedClientSuite
}

func (suite *GoClientSuite) init(address string, options ...clients.ClientOption) (context.Context, testext.SampleService) {
	return context.Background(), gen.SampleServiceClient(address, options...)
}

func (suite *GoClientSuite) ErrorMatches(err error, status int, msg string) {
	suite.Require().Error(err)
	suite.Require().Equal(status, fail.Status(err))
	suite.Require().Contains(strings.ToLower(err.Error()), strings.ToLower(msg))
}

// Ensures that we get a connection refused error when connecting to a not-running server.
func (suite *GoClientSuite) TestNotConnected() {
	ctx, client := suite.init("localhost:55555")
	_, err := client.Defaults(ctx, &testext.SampleRequest{})
	suite.ErrorMatches(err, 500, "connection refused")
}

// Ensures that we can rely on all default behaviors for an endpoint.
func (suite *GoClientSuite) TestDefaults() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address)
	res, err := client.Defaults(ctx, &testext.SampleRequest{Text: "Abide"})
	suite.Require().NoError(err)
	suite.Require().Equal("Defaults:Abide", res.Text)
}

// Ensures that we can encode/decode non-flat structs w/ nothing but strings.
func (suite *GoClientSuite) TestComplexValues() {
	address, shutdown := suite.startServer()
	defer shutdown()

	inTime := time.Date(2010, time.November, 11, 12, 0, 0, 0, time.UTC)
	inTimePtr := time.Date(2020, time.November, 11, 12, 0, 0, 0, time.UTC)

	ctx, client := suite.init(address)
	res, err := client.ComplexValues(ctx, &testext.SampleComplexRequest{
		InFlag:  true,
		InFloat: 3.14,
		InUser: testext.SampleUser{
			ID:              "abc",
			Name:            "The Dude",
			Age:             47,
			Attention:       5 * time.Second,
			AttentionString: testext.CustomDuration(4*time.Minute + 2*time.Second),
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
	suite.Require().NoError(err)
	suite.Equal(true, res.OutFlag)
	suite.Equal(3.14, res.OutFloat)
	suite.Equal("abc", res.OutUser.ID)
	suite.Equal("The Dude", res.OutUser.Name)
	suite.Equal(47, res.OutUser.Age)
	suite.Equal(5*time.Second, res.OutUser.Attention)
	suite.Equal(testext.CustomDuration(4*time.Minute+2*time.Second), res.OutUser.AttentionString)
	suite.Equal("555-1234", res.OutUser.PhoneNumber)
	suite.Equal("home@string.com", res.OutUser.MarshalToString.Home)
	suite.Equal("work@string.com", res.OutUser.MarshalToString.Work)
	suite.Equal(inTime, res.OutTime)
	suite.Require().NotNil(res.OutTimePtr)
	suite.Equal(inTimePtr, *res.OutTimePtr)

	// The Go client can use the custom (Un)MarshalJSON  implementations, so these should work.
	suite.Equal("home@object.com", res.OutUser.MarshalToObject.Home)
	suite.Equal("work@object.com", res.OutUser.MarshalToObject.Work)
}

// Ensures that we can encode/decode non-flat structs and pass those values via
// path params and query string params.
func (suite *GoClientSuite) TestComplexValuesPath() {
	address, shutdown := suite.startServer()
	defer shutdown()

	inTime := time.Date(2010, time.November, 11, 12, 0, 0, 0, time.UTC)
	inTimePtr := time.Date(2020, time.November, 11, 12, 0, 0, 0, time.UTC)

	ctx, client := suite.init(address)
	res, err := client.ComplexValuesPath(ctx, &testext.SampleComplexRequest{
		InFlag:  true,
		InFloat: 3.14,
		InUser: testext.SampleUser{
			ID:              "abc",
			Name:            "The Dude",
			Age:             47,
			Attention:       5 * time.Second,
			AttentionString: testext.CustomDuration(4*time.Minute + 2*time.Second),
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
	suite.Require().NoError(err)
	suite.Equal(true, res.OutFlag)
	suite.Equal(3.14, res.OutFloat)
	suite.Equal("abc", res.OutUser.ID)
	suite.Equal("The Dude", res.OutUser.Name)
	suite.Equal(47, res.OutUser.Age)
	suite.Equal(5*time.Second, res.OutUser.Attention)
	suite.Equal(testext.CustomDuration(4*time.Minute+2*time.Second), res.OutUser.AttentionString)
	suite.Equal("555-1234", res.OutUser.PhoneNumber)
	suite.Equal("home@string.com", res.OutUser.MarshalToString.Home)
	suite.Equal("work@string.com", res.OutUser.MarshalToString.Work)
	suite.Equal(inTime, res.OutTime)
	suite.Require().NotNil(res.OutTimePtr)
	suite.Equal(inTimePtr, *res.OutTimePtr)

	// The Go client can use the custom (Un)MarshalJSON  implementations, so these should work.
	suite.Equal("home@object.com", res.OutUser.MarshalToObject.Home)
	suite.Equal("work@object.com", res.OutUser.MarshalToObject.Work)
}

// Ensures that the client reports back 4XX style errors when they're returned.
func (suite *GoClientSuite) TestFail4XX() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address)
	_, err := client.Fail4XX(ctx, &testext.SampleRequest{})
	suite.ErrorMatches(err, 409, "always a conflict")
}

// Ensures that the client reports back 5XX style errors when they're returned.
func (suite *GoClientSuite) TestFail5XX() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address)
	_, err := client.Fail5XX(ctx, &testext.SampleRequest{})
	suite.ErrorMatches(err, 502, "always a bad gateway")
}

// Ensures that we can define a custom method/path and still send data properly.
func (suite *GoClientSuite) TestCustomRoute() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address)
	res, err := client.CustomRoute(ctx, &testext.SampleRequest{ID: "123", Text: "Abide"})
	suite.Require().NoError(err)
	suite.Equal("123", res.ID)
	suite.Equal("Route:Abide", res.Text)
}

// Ensures that we can define a custom method/path and still send data properly.
func (suite *GoClientSuite) TestCustomRouteBody() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address)
	res, err := client.CustomRouteBody(ctx, &testext.SampleRequest{ID: "123", Text: "Abide"})
	suite.Require().NoError(err)
	suite.Equal("123", res.ID)
	suite.Equal("Route:Abide", res.Text)
}

// Ensures that we can define a custom method/path and still send data properly.
func (suite *GoClientSuite) TestCustomRouteQuery() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address)
	res, err := client.CustomRouteQuery(ctx, &testext.SampleRequest{ID: "123", Text: "Abide"})
	suite.Require().NoError(err)
	suite.Equal("123", res.ID)
	suite.Equal("Route:Abide", res.Text)
}

// Ensures that the client fails if you attempt to invoke a function that has "HTTP OMIT" on it.
func (suite *GoClientSuite) TestOmitMe() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address)
	_, err := client.OmitMe(ctx, &testext.SampleRequest{})
	suite.ErrorMatches(err, 501, "")
}

// Ensures that the client can handle receiving a raw stream of data rather than auto-encoding.
func (suite *GoClientSuite) TestDownload() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address)
	res, err := client.Download(ctx, &testext.SampleDownloadRequest{Format: "text/plain"})
	suite.Require().NoError(err)
	suite.Equal(34, res.ContentLength())
	suite.Equal("text/plain", res.ContentType())
	suite.Equal("dude.txt", res.ContentFileName())

	content, _ := io.ReadAll(res.Content())
	suite.Equal("Donny, you're out of your element!", string(content))
}

// Ensures that the client can handle receiving a raw stream of data rather than auto-encoding.
func (suite *GoClientSuite) TestDownloadResumable() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address)
	res, err := client.DownloadResumable(ctx, &testext.SampleDownloadRequest{Format: "text/plain"})
	suite.Require().NoError(err)
	suite.Equal(24, res.ContentLength())
	suite.Equal("text/html", res.ContentType())
	suite.Equal("dude.html", res.ContentFileName())

	start, end, size := res.ContentRange()
	suite.Equal(50, start)
	suite.Equal(74, end)
	suite.Equal(1024, size)

	content, _ := io.ReadAll(res.Content())
	suite.Equal("<h1>The Dude Abides</h1>", string(content))
}

// Ensures that the client can handle receiving a redirect and reads it in as a raw stream.
func (suite *GoClientSuite) TestRedirect() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address)
	res, err := client.Redirect(ctx, &testext.SampleRedirectRequest{})
	suite.Require().NoError(err)
	suite.Equal(42, res.ContentLength())
	suite.Equal("text/csv", res.ContentType())

	content, _ := io.ReadAll(res.Content())
	suite.Equal("ID,Name,Enabled\n1,Dude,true\n2,Walter,false", string(content))
}

// Ensures that the client passes along authorization info specified when invoking a function.
func (suite *GoClientSuite) TestAuthorization() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address)
	ctx = metadata.WithAuthorization(ctx, "The Dude Abides")
	res, err := client.Authorization(ctx, &testext.SampleRequest{})
	suite.Require().NoError(err)
	suite.Equal("The Dude Abides", res.Text) // endpoint sets the incoming Auth value as the text
}

// Ensure that the client invocation will quit early if the underlying context is
// cancelled before the operation finishes (even if it was going to do so successfully).
func (suite *GoClientSuite) TestCancellation() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address)
	ctx, cancel := context.WithTimeout(ctx, 25*time.Millisecond)
	defer cancel()

	startTime := time.Now()
	_, err := client.Sleep(ctx, &testext.SampleRequest{})
	endTime := time.Now()

	suite.Require().Error(err)
	suite.Less(endTime.Sub(startTime), 5*time.Second, "Client seems to have waited even though context was cancelled.")
}

// Ensure that middleware put on the client fires when it's supposed to.
func (suite *GoClientSuite) TestMiddleware_order() {
	address, shutdown := suite.startServer()
	defer shutdown()

	var order []string
	ctx, client := suite.init(address, clients.WithMiddleware(
		func(request *http.Request, next clients.RoundTripperFunc) (*http.Response, error) {
			order = append(order, "A")
			res, err := next(request)
			order = append(order, "F")
			return res, err
		},
		func(request *http.Request, next clients.RoundTripperFunc) (*http.Response, error) {
			order = append(order, "B")
			res, err := next(request)
			order = append(order, "E")
			return res, err
		},
		func(request *http.Request, next clients.RoundTripperFunc) (*http.Response, error) {
			order = append(order, "C")
			res, err := next(request)
			order = append(order, "D")
			return res, err
		},
	))

	res, err := client.Defaults(ctx, &testext.SampleRequest{Text: "Dude"})
	suite.Require().NoError(err)
	suite.Equal([]string{"A", "B", "C", "D", "E", "F"}, order)
	suite.Equal("Defaults:Dude", res.Text)
}

// Ensure that middleware put on the client fires before we finalize the request to send to the server.
func (suite *GoClientSuite) TestMiddleware_modifyMetadata() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address, clients.WithMiddleware(
		func(req *http.Request, next clients.RoundTripperFunc) (*http.Response, error) {
			ctx := metadata.WithAuthorization(req.Context(), "Middleware Abides")
			return next(req.WithContext(ctx))
		},
	))

	// There's no Auth on the main context, but we shove it into the request context via middleware.
	res, err := client.Authorization(ctx, &testext.SampleRequest{})
	suite.Require().NoError(err)
	suite.Equal("Middleware Abides", res.Text)
	suite.Equal("", metadata.Authorization(ctx)) // auth only existed on the request context, not the main one
}

// Ensure that a middleware function can skip hitting the server altogether if you want.
func (suite *GoClientSuite) TestMiddleware_shortCircuit() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address, clients.WithMiddleware(
		func(req *http.Request, next clients.RoundTripperFunc) (*http.Response, error) {
			res := http.Response{
				Status:     "OK",
				StatusCode: 200,
				Header: http.Header{
					"Content-Type":   []string{"application/json"},
					"Content-Length": []string{"29"},
				},
				Body:          io.NopCloser(bytes.NewBufferString(`{"Text":"Johnny 5 is ALIVE!"}`)),
				ContentLength: 29,
				Uncompressed:  true,
			}
			return &res, nil
		},
	))

	res, err := client.Defaults(ctx, &testext.SampleRequest{Text: "Ignore This"})
	suite.Require().NoError(err)
	suite.Equal("Johnny 5 is ALIVE!", res.Text)
}

// Ensure that a middleware function can skip hitting the server altogether if you want.
func (suite *GoClientSuite) TestMiddleware_shortCircuitError() {
	address, shutdown := suite.startServer()
	defer shutdown()

	ctx, client := suite.init(address, clients.WithMiddleware(
		func(req *http.Request, next clients.RoundTripperFunc) (*http.Response, error) {
			return nil, fail.Throttled("nope")
		},
	))

	_, err := client.Defaults(ctx, &testext.SampleRequest{Text: "Ignore This"})
	suite.ErrorMatches(err, 429, "nope")
}

func (suite *GoClientSuite) TestRoles() {
	address, shutdown := suite.startServer()
	defer shutdown()
	ctx, client := suite.init(address)

	type testCase struct {
		ID       string
		UserID   string
		Expected []string
	}

	runTest := func(c testCase) {
		res, err := client.SecureWithRoles(ctx, &testext.SampleSecurityRequest{
			ID:   c.ID,
			User: testext.SampleUser{ID: c.UserID},
		})
		suite.Require().NoError(err)
		suite.Equal(c.Expected, res.Roles)
	}

	runTest(testCase{
		ID:       "",
		UserID:   "",
		Expected: []string{"admin.write", "user..write", "user..admin", "junk..crap"},
	})

	runTest(testCase{
		ID:       "123",
		UserID:   "456",
		Expected: []string{"admin.write", "user.123.write", "user.456.admin", "junk..crap"},
	})
}

func (suite *GoClientSuite) TestRolesAliased() {
	address, shutdown := suite.startServer()
	defer shutdown()
	ctx, client := suite.init(address)

	type testCase struct {
		ID       string
		FancyID  testext.StringLike
		Expected []string
	}

	runTest := func(c testCase) {
		res, err := client.SecureWithRolesAliased(ctx, &testext.SampleSecurityRequest{
			ID:      c.ID,
			FancyID: c.FancyID,
			User:    testext.SampleUser{ID: c.ID, FancyID: c.FancyID},
		})
		suite.Require().NoError(err)
		suite.Equal(c.Expected, res.Roles)
	}

	runTest(testCase{
		ID:       "FOO",
		FancyID:  "",
		Expected: []string{"admin.write", "user..write", "user..admin", "junk..crap"},
	})

	runTest(testCase{
		ID:       "FOO",
		FancyID:  "BAR",
		Expected: []string{"admin.write", "user.BAR.write", "user.BAR.admin", "junk..crap"},
	})
}

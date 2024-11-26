//go:build unit

package clients_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/bridgekit-io/frodo/internal/quiet"
	"github.com/bridgekit-io/frodo/metadata"
	"github.com/bridgekit-io/frodo/services/clients"
	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite
}

// Ensure that the default RPC client is valid.
func (suite *ClientSuite) TestNewClient_default() {
	assert := suite.Require()
	client := clients.NewClient("FooService", ":9000")

	assert.Equal("FooService", client.Name)
	assert.Equal("http://:9000", client.BaseURL)
	assert.NotNil(client.HTTP, "Default HTTP client should be non-nil")
	assert.Equal(30*time.Second, client.HTTP.Timeout, "Default HTTP client timeout should be 30 seconds")

	// Yes, you can leave these blank. You'll have a bad time... but you can do it.
	client = clients.NewClient("", "")
	assert.Equal("", client.Name)
	assert.Equal("", client.BaseURL)

	client = clients.NewClient("", "http://foo:9000/trailing/slash/")
	assert.Equal("http://foo:9000/trailing/slash", client.BaseURL, "Should trim trailing slashes in base URL")
}

// Ensure that functional options override client defaults.
func (suite *ClientSuite) TestNewClient_options() {
	assert := suite.Require()
	client := clients.NewClient("FooService", ":9000",
		func(c *clients.Client) { c.Name = "FartService" },
		func(c *clients.Client) { c.BaseURL = "https://google.com" },
		func(c *clients.Client) { c.HTTP = nil },
	)

	assert.Equal("FartService", client.Name)
	assert.Equal("https://google.com", client.BaseURL)
	assert.Nil(client.HTTP)

	httpClient := &http.Client{}
	client = clients.NewClient("FooService", ":9000", clients.WithHTTPClient(httpClient))
	assert.Same(httpClient, client.HTTP, "WithHTTPClient should set the client's HTTP client")
}

// Ensures that an RPC client can invoke an HTTP GET endpoint. All of the service request values should
// be set on the query string.
func (suite *ClientSuite) TestInvoke_get() {
	assert := suite.Require()
	client := suite.newClient(func(r *http.Request) (*http.Response, error) {
		suite.assertURL(r, "http://localhost:9000/foo")
		suite.assertQuery(r, url.Values{
			"ID":         []string{"123"},
			"Int":        []string{"42"},
			"Inner.Flag": []string{"false"}, // even includes default values for fields not explicitly set
			"Inner.Skip": []string{"100"},
		})
		return suite.respond(200, &clientResponse{ID: "Bob", Name: "Loblaw"})
	})

	in := &clientRequest{ID: "123", Int: 42, Inner: clientInner{Skip: 100}}
	out := &clientResponse{}
	err := client.Invoke(context.Background(), "GET", "/foo", in, out)
	assert.NoError(err)
	assert.Equal("Bob", out.ID)
	assert.Equal("Loblaw", out.Name)
}

// Ensures that an RPC client can invoke an HTTP POST endpoint. All of the service request values should
// be set on the body, not the query string.
func (suite *ClientSuite) TestInvoke_post() {
	assert := suite.Require()
	client := suite.newClient(func(r *http.Request) (*http.Response, error) {
		assert.Len(r.URL.Query(), 0, "Client.Invoke() - POST should not have a query string")
		suite.assertURL(r, "http://localhost:9000/foo")

		actual, err := suite.unmarshal(r)
		assert.NoError(err, "Client.Invoke() - POST should not send junk JSON")
		assert.Equal("123", actual.ID)
		assert.Equal(42, actual.Int)
		assert.Equal(100, actual.Inner.Skip)
		assert.Equal(false, actual.Inner.Flag)

		return suite.respond(200, &clientResponse{ID: "Bob", Name: "Loblaw"})
	})

	in := &clientRequest{ID: "123", Int: 42, Inner: clientInner{Skip: 100}}
	out := &clientResponse{}
	err := client.Invoke(context.Background(), "POST", "/foo", in, out)
	assert.NoError(err)
	assert.Equal("Bob", out.ID)
	assert.Equal("Loblaw", out.Name)
}

// Ensures that an RPC client fills in path params (e.g. "/{id}"->"/1234"). We will make sure
// that path param substitutions:
//
// * Param name matches request attribute exactly
// * Param name can be case-insensitive (e.g. ":id" should match field "ID")
// * Support nested params (e.g. ":Criteria.Paging.Limit")
// * Should ignore param names that don't match anything
// * If you include a request attribute in the path, do NOT include it in the query string, too.
func (suite *ClientSuite) TestInvoke_pathParams() {
	assert := suite.Require()
	client := suite.newClient(func(r *http.Request) (*http.Response, error) {
		// Missing :bar is between 42 and 100, so you'd end up with "42//100". We choose not
		// to clean this up so that it's more clear what's missing. Either way this is a
		// situation you do not want, so might as well make it obvious where the missing value is.
		// That's also why we left the trailing "/". Normally there wouldn't be one, but since
		// there's a missing path param at the end, it felt like it should be there.
		suite.assertURL(r, "http://localhost:9000/foo/123/42//100/")

		// Make sure that the path params are not in the query string, too. Values not in the path should
		// still be in there, however.
		query := r.URL.Query()
		assert.Empty(query.Get("ID"), "Should not be in the query string when field is in the path")
		assert.Empty(query.Get("Int"), "Should not be in the query string when field is in the path")
		assert.Empty(query.Get("Inner.Skip"), "Should not be in the query string when field is in the path")
		assert.Equal("true", query.Get("Inner.Flag"))
		assert.Equal("", query.Get("Inner.Test"))
		return suite.respond(200, &clientResponse{ID: "Bob", Name: "Loblaw"})
	})

	in := &clientRequest{ID: "123", Int: 42, Inner: clientInner{Skip: 100, Flag: true}}
	out := &clientResponse{}
	err := client.Invoke(context.Background(), "GET", "/foo/{ID}/{Int}/{bar}/{Inner.Skip}/{Nope}", in, out)
	assert.NoError(err)
	assert.Equal("Bob", out.ID)
	assert.Equal("Loblaw", out.Name)
}

// Ensures that an RPC client will translate 4XX/5XX errors into the
// equivalent status-coded error.
func (suite *ClientSuite) TestInvoke_httpStatusError() {
	assert := suite.Require()
	client := suite.newClient(func(r *http.Request) (*http.Response, error) {
		typeText := http.Header{"Content-Type": []string{"text/plain"}}
		typeJSON := http.Header{"Content-Type": []string{"application/json"}}
		switch r.URL.Path {
		case "/404":
			// A non-json response where the body is the error message
			body := "not here, dude"
			return &http.Response{StatusCode: 404, Header: typeText, Body: io.NopCloser(strings.NewReader(body))}, nil
		case "/409":
			body := `"already did that"`
			return &http.Response{StatusCode: 409, Header: typeJSON, Body: io.NopCloser(strings.NewReader(body))}, nil
		case "/500":
			body := `{"message": "broke as hell"}`
			return &http.Response{StatusCode: 500, Header: typeJSON, Body: io.NopCloser(strings.NewReader(body))}, nil
		case "/504":
			body := `{"foo": "broke as hell"}`
			return &http.Response{StatusCode: 504, Header: typeJSON, Body: io.NopCloser(strings.NewReader(body))}, nil
		}
		panic("how did you get here?")
	})

	out := &clientResponse{}
	err := client.Invoke(context.Background(), "POST", "/404", &clientRequest{}, out)
	assert.Error(err, "Client.Invoke() - 404 status code should return an error")
	assert.Contains(err.Error(), "not here, dude", "Client.Invoke() - should include plain text message")

	out = &clientResponse{}
	err = client.Invoke(context.Background(), "POST", "/409", &clientRequest{}, out)
	assert.Error(err, "Client.Invoke() - 409 status code should return an RPC error")
	assert.Contains(err.Error(), "already did that", "Client.Invoke() - should include json string message")

	// A json response where the body is the Responder error struct (i.e. message is a json attribute)
	out = &clientResponse{}
	err = client.Invoke(context.Background(), "POST", "/500", &clientRequest{}, out)
	assert.Error(err, "Client.Invoke() - 500 status code should return an error")
	assert.Contains(err.Error(), "broke as hell", "Client.Invoke() - should include json error struct message")

	// A json response but the body doesn't look like our normal JSON error structure.
	out = &clientResponse{}
	err = client.Invoke(context.Background(), "POST", "/504", &clientRequest{}, out)
	assert.Error(err, "Client.Invoke() - 504 status code should return an error")
	assert.NotContains(err.Error(), "broke as hell", "Client.Invoke() - not include unknown error message formats")
}

// Check all of the different ways that Invoke() can fail.
func (suite *ClientSuite) TestInvoke_roundTripError() {
	assert := suite.Require()
	client := suite.newClient(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/ok":
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{asdf}`))}, nil
		default:
			return nil, fmt.Errorf("wtf")
		}
	})

	// We failed trying to marshal the service request value as JSON
	err := client.Invoke(context.Background(), "POST", "/fail", &unableToMarshal{}, &clientResponse{})
	assert.Error(err, "Client.Invoke() should return an error when input can't be marshaled")
	assert.NotContains(err.Error(), "wtf", "Client.Invoke() error should not get to handler if failed during setup")

	// We failed creating the request to dispatch (bad http method)
	err = client.Invoke(context.Background(), "🍺", "/fail", &clientRequest{}, &clientResponse{})
	assert.Error(err, "Client.Invoke() should return an error when request can't be constructed")
	assert.NotContains(err.Error(), "wtf", "Client.Invoke() error should not get to handler if failed during setup")

	// Dispatch went ok, but the round-tripper function returned an error
	err = client.Invoke(context.Background(), "POST", "/ok", &clientRequest{}, &clientResponse{})
	assert.Error(err, "Client.Invoke() should return an error when round tripper returns an error")
	assert.NotContains(err.Error(), "wtf", "Client.Invoke() error propagate error returned by round tripper")
}

// Should invoke your middleware in the correct order before dispatching the "real" handler.
func (suite *ClientSuite) TestWithMiddleware() {
	assert := suite.Require()
	values := []string{"", "", ""}
	client := clients.NewClient("FooService", "http://localhost:9000", clients.WithMiddleware(
		func(request *http.Request, next clients.RoundTripperFunc) (*http.Response, error) {
			values[0] = "A"
			return next(request)
		},
		func(request *http.Request, next clients.RoundTripperFunc) (*http.Response, error) {
			values[1] = "B"
			return next(request)
		},
	))
	client.HTTP.Transport = clients.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		values[2] = "C"
		return suite.respond(200, &clientResponse{ID: "123"})
	})

	out := &clientResponse{}
	err := client.Invoke(context.Background(), "POST", "/foo", &clientRequest{}, out)
	assert.NoError(err)
	assert.Equal("123", out.ID)
	assert.Equal("A", values[0])
	assert.Equal("B", values[1])
	assert.Equal("C", values[2])
}

func (suite *ClientSuite) TestInvoke_includeHeaders() {
	assert := suite.Require()
	client := suite.newClient(func(r *http.Request) (*http.Response, error) {
		assert.Equal("Hello", r.Header.Get("Authorization"))
		assert.Equal(`{"Authorization":"Hello","Values":{"Foo":{"value":"Bar"}}}`, r.Header.Get("X-RPC-Metadata"))
		return suite.respond(200, &clientResponse{ID: "123"})
	})

	ctx := context.Background()
	ctx = metadata.WithAuthorization(ctx, "Hello")
	ctx = metadata.WithValue(ctx, "Foo", "Bar")
	assert.NoError(client.Invoke(ctx, "POST", "/foo",
		&clientRequest{},
		&clientResponse{},
	))
}

func (suite *ClientSuite) newClient(roundTripper clients.RoundTripperFunc) clients.Client {
	client := clients.NewClient("Test", "http://localhost:9000")
	client.HTTP.Transport = roundTripper
	return client
}

func (suite *ClientSuite) assertURL(r *http.Request, expected string) {
	actual, _, _ := strings.Cut(r.URL.String(), "?")
	suite.Require().Equal(expected, actual, "Client.Invoke() - Incorrect URL/path")
}

func (suite *ClientSuite) assertQuery(r *http.Request, expectedValues url.Values) {
	actual := r.URL.Query()
	for key, expected := range expectedValues {
		suite.Require().ElementsMatch(expected, actual[key], "Client.Invoke() - wrong query string value for %s", key)
	}
}

func (suite *ClientSuite) respond(status int, body *clientResponse) (*http.Response, error) {
	jsonBytes, _ := json.Marshal(body)
	jsonString := strings.NewReader(string(jsonBytes))
	return &http.Response{StatusCode: status, Body: io.NopCloser(jsonString)}, nil
}

func (suite *ClientSuite) unmarshal(r *http.Request) (*clientRequest, error) {
	defer quiet.Close(r.Body)
	out := &clientRequest{}
	return out, json.NewDecoder(r.Body).Decode(out)
}

type clientRequest struct {
	ID       string
	Int      int
	Inner    clientInner
	InnerPtr *clientInner
}

type clientInner struct {
	Test string
	Flag bool
	Skip int
}

type clientResponse struct {
	ID   string
	Name string
}

type unableToMarshal struct {
	Channel chan string
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}

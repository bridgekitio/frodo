//go:build integration

package generate_test

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
	"time"

	"github.com/bridgekitio/frodo/fail"
	"github.com/bridgekitio/frodo/internal/testext"
	gen "github.com/bridgekitio/frodo/internal/testext/gen"
	"github.com/bridgekitio/frodo/services"
	"github.com/bridgekitio/frodo/services/gateways/apis"
	"github.com/stretchr/testify/suite"
)

// GeneratedClientSuite is a test suite that validates the behavior of service clients generated for languages other
// than Go. It relies on you having a "runner" executable in the client's target language that runs one of the test
// cases we want to test and parses stdout of that program to determine whether the test should pass/fail.
//
// The suite contains the logic to fire up a local instance of the gateway for the client to hit on the desired port
// as well as the ability to shut it down after the test. You can then analyze each line of stdout to determine if
// each interaction behaved as expected and write your Go assertions based on that. There will be more detail in the
// Frodo architecture Markdown docs as to how this all works.
type GeneratedClientSuite struct {
	suite.Suite
	addresses testext.FreeAddress
}

func (suite *GeneratedClientSuite) startServer() (string, func()) {
	address := suite.addresses.Next()
	sequence := &testext.Sequence{}
	server := services.NewServer(
		services.Listen(apis.NewGateway(address)),
		services.Register(gen.SampleServiceServer(testext.SampleServiceHandler{Sequence: sequence})),
	)
	go func() { _ = server.Run() }()

	// Kinda crappy, but we need some time to make sure the server is up. Sometimes
	// this goes so fast that the test case fires before the server is fully running.
	// As a result the cases fail because the server's not running... duh.
	time.Sleep(25 * time.Millisecond)

	return address, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}
}

// ExpectPass analyzes line N (zero-based) of the client runner's output and asserts that it was a successful
// interaction (e.g. output was "OK { ... some json ... }"). It will decode the right-hand-side JSON onto your
// 'out' parameter so that you can also run custom checks to fire after the decoding is complete to ensure that
// the actual output value is what you expect.
//
//	response := calc.AddResponse{}
//	suite.ExpectPass(output, 0, &response, func() {
//	    suite.Equal(12, response.Result)
//	})
func (suite *GeneratedClientSuite) ExpectPass(result ClientTestResult, out any, additionalChecks ...func()) {
	suite.Require().Equal(true, result.Pass, "Line %d: Failed when it should have passed: %v", result.Index, result.String())
	suite.Require().NoError(result.Decode(out), "Line %d: Unable to unmarshal JSON", result.Index)
	for _, additionalCheck := range additionalChecks {
		additionalCheck()
	}
}

// ExpectErrorMatching analyzes line N (zero-based) of the client runner's output and asserts that it was a failed
// interaction (e.g. output was "FAIL { ... some json ... }"). If the error conforms to our standard fail.StatusError
// type, we'll compare the status code to it. We'll check to see that msg appears ANYWHERE within the error message.
//
//	// Assumes the first case failed w/ a 403 status, the second w/ a 502, and the last with a 500.
//	suite.ExpectFail(output[0], 403, "forbidden")
//	suite.ExpectFail(output[1], 502, "bad gateway")
//	suite.ExpectFail(output[2], 500, "wtf")
func (suite *GeneratedClientSuite) ExpectFail(result ClientTestResult, status int, msg string) {
	err := fail.StatusError{}

	// If the failure string matches our {"Status":404, "Message":"Not Found"} format, decode it
	// onto a failure. If it doesn't just set the whole error string to the Message field.
	if decodingErr := result.Decode(&err); decodingErr != nil {
		err.Message = string(result.Output)
	}

	suite.Require().Equal(false, result.Pass, "Line %d: Passed when it should have failed: %s", result.Index, result.String())
	suite.Equal(status, err.StatusCode())
	suite.Contains(strings.ToLower(err.Error()), strings.ToLower(msg))
}

// RunExternalTest executes the language-specific runner to execute a single test case in that language. The result
// of the runner's execution are written to stdout w/ each interaction on a separate line. This will delegate to
// ParseClientTestOutput() to turn it into an easily workable value that you can hit with your Go assertions.
//
// If the command fails to execute properly (shell command fails entirely, not that it outputs failure
// test case stuff), this will halt the current test case right here and now.
func (suite *GeneratedClientSuite) RunExternalTest(command string) ClientTestResults {
	stdout, err := exec.Command("/bin/zsh", "-c", command).Output()
	var stderr []byte

	var execErr *exec.ExitError
	if errors.As(err, &execErr) {
		stderr = execErr.Stderr
	}

	suite.Require().NoError(err, "Running '%v' should not fail at all: [error=%v] [stdout=%s] [stderr=%s]", command, err, string(stdout), string(stderr))
	return ParseClientTestOutput(stdout)
}

// ParseClientTestOutput accepts the entire stdout of RunClientTest and parses each line to determine how each
// interaction in the test case behaved. Here is a sample output of a runner that performed 5 client calls to
// the backend service; 3 that passed and 2 that failed.
//
//	OK {"result": 5}
//	FAIL {"message": "divide by zero", "status": 400}
//	OK {"result": 3.14}
//	OK {"result": 10, "remainder": 2}
//	FAIL {"message": "overflow", "status": 400}
//
// All language runners should output in this format for this to work. It's a convention that allows us to build
// assertions in Go regardless of how the target language does its work.
func ParseClientTestOutput(stdout []byte) ClientTestResults {
	var results ClientTestResults

	for i, line := range strings.Split(strings.TrimSpace(string(stdout)), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
			// Ignore blank lines
		case strings.HasPrefix(line, "OK "):
			results = append(results, ClientTestResult{
				Index:  i,
				Pass:   true,
				Output: []byte(strings.TrimSpace(strings.TrimPrefix(line, "OK ")))})

		case strings.HasPrefix(line, "FAIL "):
			results = append(results, ClientTestResult{
				Index:  i,
				Pass:   false,
				Output: []byte(strings.TrimSpace(strings.TrimPrefix(line, "FAIL ")))})

		default:
			// you probably just wrote a debugging line... ignore it.
		}
	}
	return results
}

// ClientTestResults encapsulates all of the output lines from a client test runner.
type ClientTestResults []ClientTestResult

// ClientTestResult decodes a single output line of stdout from a client test runner. It parses "OK {...}"
// or "FAIL {...}" and makes it easier for your test assertion code to work with.
type ClientTestResult struct {
	// Index is the 0-based line index of where this result appeared in the output.
	Index int
	// Pass is true when the line started with "OK", false otherwise.
	Pass bool
	// Output is the raw characters output to stdout for this interaction.
	Output []byte
}

// Decode overlays this output line's JSON on top of the 'out' parameter (i.e. the stuff after OK/FAIL).
func (res ClientTestResult) Decode(out any) error {
	return json.Unmarshal(res.Output, out)
}

// String just regurgitates the original output line.
func (res ClientTestResult) String() string {
	return string(res.Output)
}

// RawClientOutput matches the data structure of the Node/JS object returned by service
// functions that result in "raw" byte responses.
type RawClientOutput struct {
	// Content contains the raw byte content output by the service call.
	Content string
	// ContentType contains the captured "Content-Type" header data.
	ContentType string
	// ContentFileName contains the captured file name from the "Content-Disposition" header data.
	ContentFileName string
	// ContentLength captures the Content-Length header info.
	ContentLength int
	// ContentRange contains the captured Content-Range header info.
	ContentRange struct {
		// Unit is probably the value 'bytes'
		Unit string
		// Start is the starting index ("X" in "bytes X-Y/Z")
		Start int
		// End is the ending index ("Y" in "bytes X-Y/Z")
		End int
		// Size number of bytes in the whole resource ("Z" in "bytes X-Y/Z")
		Size int
	}
}

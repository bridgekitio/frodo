# Since our suites run as a single Go test (testify just makes it look like different
# tests), this timeout should be long enough to handle our slowest suite.
TEST_TIMEOUT=60s
VERBOSE=false

# Let's us optionally enable verbose output when running tests.
ifeq ($(VERBOSE),true)
	GO_VERBOSE_FLAG=-v
else
	GO_VERBOSE_FLAG=
endif

#
# Builds the actual frodo CLI executable.
#
build:
	@ \
	go build -o out/frodo main.go

install: build
	@ \
 	echo "Overwriting go-installed version..." && \
 	cp out/frodo $$GOPATH/bin/frodo

#
# Runs the all of the test suites for the entire Frodo module.
#
test: test-unit test-integration

#
# Runs the self-contained unit tests that don't require code generation or anything like that to run.
#
test-unit:
	@ \
	go test $(GO_VERBOSE_FLAG) -count=1 -timeout $(TEST_TIMEOUT) -tags unit ./...

#
# Generates the clients for all of our supported languages (Go, JS, Dart) and runs tests on them
# to make sure that they all behave as expected. So not only can we generate them, but can we actually
# fetch data from the sample service and get the expected results back?
#
test-integration: generate
	@ \
	go test $(GO_VERBOSE_FLAG) -count=1 -timeout $(TEST_TIMEOUT) -tags integration ./...

#
# Runs the go:generate utility on all of our sample services that we use in unit/integration tests.
#
generate: build
	@ \
	go generate ./internal/testext/... && \
	mv ./internal/testext/gen/*.client.js ./generate/testdata/js/ && \
	mv ./internal/testext/gen/*.client.dart ./generate/testdata/dart/

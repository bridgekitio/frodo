ROOT_DIR=../..

#
# This runs the calculator service in a local process exposing it on port 9000.
#
server: generate
	@ go build -o out/server cmd/main.go && out/server

#
# This runs the sample consumer/client program that utilizes the service client to make
# requests to the calculator service. This will fail if you do not have 'run-server'
# already running.
#
client:
	@ go build -o out/client cmd/client/main.go && out/client

#
# Generates all of our RPC gateway/client artifacts from the service definitions.
#
generate: build-generator
	@ \
 	$(ROOT_DIR)/out/frodo server calc/calc_service.go && \
 	$(ROOT_DIR)/out/frodo client calc/calc_service.go && \
	$(ROOT_DIR)/out/frodo client calc/calc_service.go --language=node


build-generator:
	@ \
 	cd $(ROOT_DIR) && \
 	$(MAKE) build

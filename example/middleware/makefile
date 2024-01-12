ROOT_DIR=../..

#
# This runs the calculator service in a local process exposing it on port 9000.
#
server: generate
	@ go build -o out/server cmd/main.go && out/server

#
# Generates all of our RPC gateway/client artifacts from the service definitions.
#
generate: build-generator
	@ \
 	$(ROOT_DIR)/out/frodo server ../basic/calc/calc_service.go && \
 	$(ROOT_DIR)/out/frodo client ../basic/calc/calc_service.go


build-generator:
	@ \
 	cd $(ROOT_DIR) && \
 	$(MAKE) build

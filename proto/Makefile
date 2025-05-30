ifeq (/,${HOME})
GOLANGCI_LINT_CACHE=/tmp/golangci-lint-cache/
else
GOLANGCI_LINT_CACHE=${HOME}/.cache/golangci-lint
endif
GOLANGCI_LINT ?= GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) go run github.com/golangci/golangci-lint/cmd/golangci-lint

include tools/tools.mk

check-gopath:
ifndef GOPATH
  $(error GOPATH is undefined)
endif

# .PHONY: generate generate-proto generate-mocks
# generate: generate-proto generate-mocks

PROTO_PREFIX:=github.com/nitrictech/nitric/proto

clean-proto:
	@rm -rf ./pkg/proto/

.PHONY: generate-proto-go generate-proto-pythons

# Generate interfaces
generate-proto-go: install-tools check-gopath clean-proto
	@echo Generating Proto Sources
	@$(PROTOC) --go_out=. --go_opt=module=$(PROTO_PREFIX) --go-grpc_opt=module=$(PROTO_PREFIX) --go-grpc_out=. ./*/v2/*.proto -I .

PYTHON_SDK_DIR=../cli/pkg/sdk/

generate-proto-python: install-tools clean-proto
	@echo Generating Proto Sources
	@rm -rf ${PYTHON_SDK_DIR}/python_proto
	@python3 -m grpc_tools.protoc -I . --python_betterproto_out=${PYTHON_SDK_DIR} ./*/v2/*.proto
	@mv ${PYTHON_SDK_DIR}/nitric/proto ${PYTHON_SDK_DIR}/python_proto
	@rm -rf ${PYTHON_SDK_DIR}/nitric
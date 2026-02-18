# ==============================================================================
# gemara-content-service Makefile
# ==============================================================================
#
# Usage:
#   make all         - Runs tests and then builds the binary
#   make test        - Runs tests with coverage
#   make build       - Builds the binary and places it in the ./bin directory
#   make clean       - Removes generated binaries and build artifacts
#   make help        - Displays this help message
# ==============================================================================

BIN_DIR := bin
BINARY := compass
CERT_DIR := hack/self-signed-cert
OPENSSL_CNF := $(CERT_DIR)/openssl.cnf

all: test build

# ------------------------------------------------------------------------------
# Test
# ------------------------------------------------------------------------------
test: ## Runs unit tests with coverage
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | tail -n1
.PHONY: test

test-race: ## Runs tests with race detection
	go test -v -race ./...
.PHONY: test-race

coverage-report: test ## Generate HTML coverage report and show summary
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"
.PHONY: coverage-report

# ------------------------------------------------------------------------------
# Build
# ------------------------------------------------------------------------------
build: ## Builds the binary and places it in the $(BIN_DIR) directory
	@mkdir -p $(BIN_DIR)
	go build -v -o $(BIN_DIR)/$(BINARY) ./cmd/compass/
	@echo "--- Binary built: $(BIN_DIR)/$(BINARY) ---"
.PHONY: build

clean: ## Removes all generated binaries and build artifacts
	@echo "--- Cleaning up build artifacts ---"
	@rm -rf $(BIN_DIR) coverage.out coverage.html
	@echo "--- Cleanup complete ---"
.PHONY: clean

# ------------------------------------------------------------------------------
# Dependencies
# ------------------------------------------------------------------------------
deps: ## Tidy, verify, and download dependencies
	go mod tidy
	go mod verify
	go mod download
.PHONY: deps

# ------------------------------------------------------------------------------
# Code Generation
# ------------------------------------------------------------------------------
api-codegen: ## Runs go generate for OpenAPI code generation
	go generate ./...
.PHONY: api-codegen

# ------------------------------------------------------------------------------
# Certificates
# ------------------------------------------------------------------------------
generate-self-signed-cert: ## Generate self-signed certificates for local development
	@find $(CERT_DIR) -mindepth 1 ! -name 'openssl.cnf' -delete
	@echo "--- Generating self-signed certificates in $(CERT_DIR) ---"
	@openssl genrsa -out $(CERT_DIR)/ca.key 2048
	@openssl req -x509 -new -nodes -key $(CERT_DIR)/ca.key -sha256 -days 365 \
		-subj "/CN=Gemara Content Service CA" \
		-extensions v3_ca -config $(OPENSSL_CNF) \
		-out $(CERT_DIR)/ca.crt
	@openssl genrsa -out $(CERT_DIR)/compass.key 2048
	@chmod a+r $(CERT_DIR)/compass.key
	@openssl req -new -key $(CERT_DIR)/compass.key -out $(CERT_DIR)/compass.csr -config $(OPENSSL_CNF)
	@openssl x509 -req -in $(CERT_DIR)/compass.csr -CA $(CERT_DIR)/ca.crt -CAkey $(CERT_DIR)/ca.key -CAcreateserial \
		-out $(CERT_DIR)/compass.crt -days 365 -sha256 \
		-extfile $(OPENSSL_CNF) -extensions v3_req
	@echo "--- Certificates generated successfully ---"
.PHONY: generate-self-signed-cert

# ------------------------------------------------------------------------------
# Linting
# ------------------------------------------------------------------------------
golangci-lint: ## Runs golangci-lint
	golangci-lint run ./...
.PHONY: golangci-lint

# ------------------------------------------------------------------------------
# Help
# ------------------------------------------------------------------------------
help: ## Display this help screen
	@grep -E '^[a-z.A-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

.PHONY: build run test clean fmt vet generate check-generate help 

# Go binary path
GOBIN := $(shell go env GOPATH)/bin

# Build the application
build:
	go build -o bin/service-provider-api ./cmd/service-provider-api

# Check AEP compliance
aep:
	spectral lint .spectral.yaml ./api/v1alpha1/openapi.yaml

# Run the application
run:
	go run ./cmd/service-provider-api

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Install dependencies
tidy:
	go mod tidy

# Run all checks
check: fmt vet test

# Build and run
dev: build
	./bin/service-provider-api

##################### "make generate" support start ##########################
MOQ := $(GOBIN)/moq

# Install moq if not already present
$(MOQ):
	@echo "📦 Installing moq..."
	@go install github.com/matryer/moq@latest
	@echo "✅ 'moq' installed successfully."

# Code generation
generate: $(MOQ)
	@echo "⚙️ Running go generate..."
	@PATH="$(GOBIN):$$PATH" go generate -v $(shell go list ./...)
	@echo "⚙️ Running mockgen script..."
	@hack/mockgen.sh
	@$(MAKE) format
	@echo "✅ Generate complete."

# Check if generate changes the repo
check-generate: generate
	@echo "🔍 Checking if generated files are up to date..."
	@git diff --quiet || (echo "❌ Detected uncommitted changes after generate. Run 'make generate' and commit the result." && git status && exit 1)
	@echo "✅ All generated files are up to date."
##################### "make generate" support end   ##########################

##################### "make format" support start ##########################
GOIMPORTS := $(GOBIN)/goimports

# Install goimports if not already available
$(GOIMPORTS):
	@echo "📦 Installing goimports..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ 'goimports' installed successfully."

# Format Go code using gofmt and goimports
format: $(GOIMPORTS)
	@echo "🧹 Formatting Go code..."
	@gofmt -s -w .
	@$(GOIMPORTS) -w .
	@echo "✅ Format complete."

# Check that formatting does not introduce changes
check-format: format
	@echo "🔍 Checking if formatting is up to date..."
	@git diff --quiet || (echo "❌ Detected uncommitted changes after format. Run 'make format' and commit the result." && git status && exit 1)
	@echo "✅ All formatted files are up to date."
##################### "make format" support end   ##########################

# Help
help:
	@echo "Available targets:"
	@echo "  build           - Build the application"
	@echo "  run             - Run the application"
	@echo "  test            - Run tests"
	@echo "  clean           - Clean build artifacts"
	@echo "  fmt             - Format code"
	@echo "  vet             - Vet code"
	@echo "  tidy            - Tidy dependencies"
	@echo "  check           - Run all checks (fmt, vet, test)"
	@echo "  dev             - Build and run"
	@echo "  generate        - Generate code from OpenAPI specification"
	@echo "  help            - Show this help"

include deploy/deploy.mk

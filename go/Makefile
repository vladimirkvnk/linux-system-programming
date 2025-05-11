.PHONY: lint test

# Install golangci-lint
GOLANGCI_LINT_VERSION = v2.1.2
GOLANGCI_LINT = $(shell go env GOPATH)/bin/golangci-lint

$(GOLANGCI_LINT):
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

# Sync workspace
sync:
	go work sync

# Lint all modules
lint: $(GOLANGCI_LINT)
	find . -name go.mod -not -path "./go.mod" | while read modfile; do \
		dir=$$(dirname "$$modfile"); \
		echo "🛠️ Linting $$dir"; \
		(cd "$$dir" && golangci-lint run ./...); \
	done

# Test all modules
test:
	find . -name go.mod -not -path "./go.mod" | while read modfile; do \
		dir=$$(dirname "$$modfile"); \
		echo "🧪 Testing $$dir"; \
		(cd "$$dir" && go test -v -race ./...); \
	done

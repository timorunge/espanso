.DEFAULT_GOAL := help

.PHONY: fmt fmt-fix tidy vet lint test test-ci check help

## fmt: Check formatting (fails on diff)
fmt:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Unformatted files:"; echo "$$unformatted"; exit 1; \
	fi

## fmt-fix: Fix Go formatting (destructive)
fmt-fix:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Fixing unformatted files..."; \
		gofmt -s -w .; \
	fi

## tidy: Check that go.mod and go.sum are tidy
tidy:
	go mod tidy -diff

## vet: Run go vet
vet:
	go vet ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## test: Run tests (short mode, race, 2m timeout)
test:
	go test -race -short -timeout 2m ./...

## test-ci: Run tests (full, race, 5m timeout)
test-ci:
	go test -race -timeout 5m ./...

## check: Run all quality gates (fmt tidy vet lint test)
check: fmt tidy vet lint test

## help: Show this help
help:
	@echo "espanso Makefile"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'

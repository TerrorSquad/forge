BINARY     := forge
INSTALL    := $(HOME)/.local/bin/$(BINARY)
BUILD_FLAGS := -ldflags="-s -w" -trimpath
CGO_ENABLED := 0

.PHONY: build install test clean coverage

build:
	CGO_ENABLED=$(CGO_ENABLED) go build $(BUILD_FLAGS) -o $(BINARY) ./cmd/forge

install: build
	cp $(BINARY) $(INSTALL)

test:
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Generated coverage: coverage.out"

coverage:
	@go tool cover -func=coverage.out

clean:
	rm -f $(BINARY)

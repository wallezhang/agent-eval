.PHONY: build test vet lint clean run license-check license-fix

BINARY := agent-eval
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test ./... -v

vet:
	go vet ./...

lint: vet
	@which golangci-lint > /dev/null 2>&1 || echo "golangci-lint not installed"
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run ./...

clean:
	rm -f $(BINARY)
	rm -rf results/

run:
	go run . run -c examples/simple/eval.yaml

install:
	go install $(LDFLAGS) .

LICENSE_HEADER_LINE1 := // Copyright 2026 wallezhang. All rights reserved.
LICENSE_HEADER_LINE2 := // SPDX-License-Identifier: Apache-2.0

license-check:
	@missing=0; \
	for f in $$(find . -name '*.go'); do \
		head -1 "$$f" | grep -qF '$(LICENSE_HEADER_LINE1)' && \
		sed -n '2p' "$$f" | grep -qF '$(LICENSE_HEADER_LINE2)' || \
		{ echo "$$f"; missing=1; }; \
	done; \
	if [ $$missing -eq 1 ]; then echo "Missing license header in above files"; exit 1; fi; \
	echo "All Go files have license headers"

license-fix:
	@for f in $$(find . -name '*.go'); do \
		head -1 "$$f" | grep -qF '$(LICENSE_HEADER_LINE1)' && \
		sed -n '2p' "$$f" | grep -qF '$(LICENSE_HEADER_LINE2)' || \
		{ printf '%s\n%s\n\n' '$(LICENSE_HEADER_LINE1)' '$(LICENSE_HEADER_LINE2)' | cat - "$$f" > "$$f.tmp" && mv "$$f.tmp" "$$f"; echo "Fixed: $$f"; }; \
	done; \
	echo "Done"

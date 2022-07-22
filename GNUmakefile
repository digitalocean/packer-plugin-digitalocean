NAME=digitalocean
BINARY=packer-plugin-${NAME}

COUNT?=1
TEST?=$(shell go list ./...)
HASHICORP_PACKER_PLUGIN_SDK_VERSION?=$(shell go list -m github.com/hashicorp/packer-plugin-sdk | cut -d " " -f2)

.PHONY: dev

build:
	@go build -o ${BINARY}

dev: build
	@mkdir -p ~/.packer.d/plugins/
	@mv ${BINARY} ~/.packer.d/plugins/${BINARY}

test:
	@go test -v -race -count $(COUNT) $(TEST) -timeout=3m

install-packer-sdc: ## Install packer sofware development command
	@go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@${HASHICORP_PACKER_PLUGIN_SDK_VERSION}

ci-release-docs: install-packer-sdc
	@packer-sdc renderdocs -src docs -partials docs-partials/ -dst docs/
	@/bin/sh -c "[ -d docs ] && zip -r docs.zip docs/"

plugin-check: install-packer-sdc build
	@packer-sdc plugin-check ${BINARY}

testacc: dev
	@PACKER_ACC=1 go test -count $(COUNT) -v $(TEST) -timeout=120m

generate: install-packer-sdc
	@go generate ./...

check-generate: generate
	echo "==> Checking that auto-generated code is not changed..."
	git diff --exit-code; if [ $$? -eq 1 ]; then \
		echo "Found diffs in go generated code."; \
		echo "You can use the command: \`make generate\` to regenerate code."; \
		exit 1; \
	fi

install-golangci-lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint: install-golangci-lint
	@golangci-lint run ./...

fmt:
	@go fmt ./...

check-fmt: fmt
	echo "==> Checking that code complies with go fmt requirements..."
	git diff --exit-code; if [ $$? -eq 1 ]; then \
		echo "Found files that are not fmt'ed."; \
		echo "You can use the command: \`go fmt ./...\` to reformat code."; \
		exit 1; \
	fi
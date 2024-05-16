NAME=digitalocean
BINARY=packer-plugin-${NAME}
PLUGIN_FQN="$(shell grep -E '^module' <go.mod | sed -E 's/module *//')"

COUNT?=1
TEST?=$(shell go list ./...)
HASHICORP_PACKER_PLUGIN_SDK_VERSION?=$(shell go list -m github.com/hashicorp/packer-plugin-sdk | cut -d " " -f2)

.PHONY: dev

build:
	@go build -o ${BINARY}

dev:
	@go build -ldflags="-X '${PLUGIN_FQN}/version.VersionPrerelease=dev'" -o '${BINARY}'
	packer plugins install --path ${BINARY} "$(shell echo "${PLUGIN_FQN}" | sed 's/packer-plugin-//')"

test:
	@go test -v -race -count $(COUNT) ./...  $(TEST) -timeout=3m

install-packer-sdc: ## Install packer sofware development command
	@go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@${HASHICORP_PACKER_PLUGIN_SDK_VERSION}

plugin-check: install-packer-sdc build
	@packer-sdc plugin-check ${BINARY}

testacc: dev
	@PACKER_ACC=1 go test -count $(COUNT) -v ./...  $(TEST) -timeout=120m

generate: install-packer-sdc
	@go generate ./...
	@if [ -d ".docs" ]; then rm -r ".docs"; fi
	@packer-sdc renderdocs -src "docs" -partials docs-partials/ -dst ".docs/"
	@./.web-docs/scripts/compile-to-webdocs.sh "." ".docs" ".web-docs" "digitalocean"
	@rm -r ".docs"

check-generate: generate
	echo "==> Checking that auto-generated code is not changed..."
	git diff --exit-code; if [ $$? -eq 1 ]; then \
		echo "Found diffs in go generated code."; \
		echo "You can use the command: \`make generate\` to regenerate code."; \
		exit 1; \
	fi

install-golangci-lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.2

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

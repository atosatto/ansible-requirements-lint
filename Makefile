# Name of the package and tool
NAME := ansible-requirements-lint
PKG := github.com/atosatto/$(NAME)

# Set the dir where built cross-compiled binaries will be output
BUILDDIR := $(shell pwd)/bin

# Populate version variables
VERSION := $(shell git describe $(git rev-list --tags --max-count=1))
GITCOMMIT := $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifeq ($(VERSION),)
	VERSION := 0.0.0
endif
ifneq ($(GITUNTRACKEDCHANGES),)
	VERSION := $(GITCOMMIT)-dirty
endif

# Go compiler name
GO := go

# Go compiler and linker flags
LDFLAGS := "-s -w -X main.version=$(VERSION)"

all: clean fmt lint test staticcheck vet build ## Lints, tests and builds the code.

.PHONY: build
build: $(NAME) ## Builds a dynamic executable or package.

$(NAME): $(shell find . -type f -name '*.go') Makefile
	$(GO) build -ldflags $(LDFLAGS) -o $(BUILDDIR)/$(NAME) ./cmd/$(NAME)

.PHONY: fmt
fmt: ## Verifies that all files are `gofmt`ed.
	@if [[ ! -z "$(shell gofmt -s -l . | grep -v vendor | tee /dev/stderr)" ]]; then \
		exit 1; \
	fi

.PHONY: lint
lint: ## Verifies that `golint` passes.
	@if [[ ! -z "$(shell golint ./... | grep -v vendor | tee /dev/stderr)" ]]; then \
		exit 1; \
	fi

.PHONY: vet
vet: ## Verifies that `go vet` passes.
	@if [[ ! -z "$(shell $(GO) vet $(shell $(GO) list ./... | grep -v vendor) | tee /dev/stderr)" ]]; then \
		exit 1; \
	fi

.PHONY: staticcheck
staticcheck: ## Verifies that `staticcheck` passes.
	@if [[ ! -z "$(shell staticcheck $(shell $(GO) list ./... | grep -v vendor) | tee /dev/stderr)" ]]; then \
		exit 1; \
	fi

.PHONY: test
test: ## Runs the go tests.
	@$(GO) test -v $(shell $(GO) list ./... | grep -v vendor)

.PHONY: coverage
coverage: ## Runs the go test and builds a coverage report.
	$(GO) test -race -coverprofile=coverage.out -covermode=atomic $(shell $(GO) list ./... | grep -v vendor)
	$(GO) tool cover -html=coverage.out -o coverage.html

.PHONY: install
install: ## Installs the binaries.
	$(GO) install -a -ldflags $(LDFLAGS) ./cmd/$(NAME)

.PHONY: vendor
vendor: ## Updates the vendors directory.
	@$(RM) go.sum
	@$(RM) -r vendor
	GO111MODULE=on $(GO) mod init || true
	GO111MODULE=on $(GO) mod tidy
	GO111MODULE=on $(GO) mod vendor

.PHONY: clean
clean: ## Cleanup all the build artifacts.
	$(RM) -r $(BUILDDIR)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | sed 's/:.*##/: ##/g' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Abort makefile if variable not set
check_defined = \
	$(strip $(foreach 1,$1, \
		$(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
	$(if $(value $1),, \
		$(error Undefined $1$(if $2, ($2))))

#   Copyright IBM Corporation 2022
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#        http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

GO_VERSION  ?= $(shell go run ./scripts/detectgoversion/detect.go 2>/dev/null || printf '1.18')
BINNAME     ?= konveyor
BINDIR      := $(CURDIR)/bin
DISTDIR		:= $(CURDIR)/_dist
TARGETS     := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64
MULTI_ARCH_TARGET_PLATFORMS := linux/amd64,linux/arm64
REGISTRYNS  ?= quay.io/konveyor

GOPATH        = $(shell go env GOPATH)
GOX           = $(GOPATH)/bin/gox
GOTEST        = ${GOPATH}/bin/gotest
GOLANGCILINT  = $(GOPATH)/bin/golangci-lint 
GOLANGCOVER   = $(GOPATH)/bin/goveralls 

PKG        := ./...
LDFLAGS    := -w -s

SRC        = $(shell find . -type f -name '*.go' -print)
ARCH       = $(shell uname -p)
GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git tag --points-at | tail -n 1)
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

GOGET     := cd / && GO111MODULE=on go install 

ifdef VERSION
	BINARY_VERSION = $(VERSION)
endif
BINARY_VERSION ?= ${GIT_TAG}
ifneq ($(BINARY_VERSION),)
	LDFLAGS += -X github.com/konveyor/cli/lib.version=${BINARY_VERSION}
	VERSION ?= $(BINARY_VERSION)
endif
VERSION ?= latest

VERSION_METADATA = unreleased
ifneq ($(GIT_TAG),)
	VERSION_METADATA =
endif
LDFLAGS += -X github.com/konveyor/cli/lib.buildmetadata=${VERSION_METADATA}

LDFLAGS += -X github.com/konveyor/cli/lib.gitCommit=${GIT_COMMIT}
LDFLAGS += -X github.com/konveyor/cli/lib.gitTreeState=${GIT_DIRTY}
# LDFLAGS += -extldflags "-static"

# HELP
# This will output the help for each task
.PHONY: help
help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[0-9a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# -- Build --

.PHONY: build
build: get $(BINDIR)/$(BINNAME) ## Build go code
	@printf "\033[32m-------------------------------------\n BUILD SUCCESS\n-------------------------------------\033[0m\n"

$(BINDIR)/$(BINNAME): $(SRC) $(ASSETS) $(WEB_ASSETS)
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(BINNAME) .
	mkdir -p $(GOPATH)/bin/
	cp $(BINDIR)/$(BINNAME) $(GOPATH)/bin/

.PHONY: get
get: go.mod
	go mod download

.PHONY: generate
generate: ## Generate assets
	go generate ${PKG}

# -- Test --

.PHONY: test
test: ## Run tests
	go test -run . $(PKG) -race
	@printf "\033[32m-------------------------------------\n TESTS PASSED\n-------------------------------------\033[0m\n"

${GOTEST}:
	${GOGET} github.com/rakyll/gotest@v0.0.6

.PHONY: test-verbose
test-verbose: ${GOTEST}
	gotest -run . $(PKG) -race -v

${GOLANGCOVER}:
	${GOGET} github.com/mattn/goveralls@v0.0.11

.PHONY: test-coverage
test-coverage: ${GOLANGCOVER} ## Run tests with coverage
	go test -run . $(PKG) -coverprofile=coverage.txt -covermode=atomic

${GOLANGCILINT}:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.45.2

.PHONY: test-style
test-style: ${GOLANGCILINT} 
	${GOLANGCILINT} run --timeout 3m
	scripts/licensecheck.sh
	@printf "\033[32m-------------------------------------\n STYLE CHECK PASSED\n-------------------------------------\033[0m\n"

# -- CI --

.PHONY: ci
ci: clean build test test-style ## Run CI routine

# -- Release --

$(GOX):
	${GOGET} github.com/mitchellh/gox@v1.0.1

.PHONY: build-cross
build-cross: $(GOX) clean $(SRC) $(ASSETS) $(WEB_ASSETS)
	CGO_ENABLED=0 $(GOX) -parallel=3 -output="$(DISTDIR)/{{.OS}}-{{.Arch}}/$(BINNAME)" -osarch='$(TARGETS)' -ldflags '$(LDFLAGS)' .

.PHONY: dist
dist: clean build-cross ## Build distribution
	mkdir -p $(DISTDIR)/files
	cp -r ./LICENSE $(DISTDIR)/files/
	cd $(DISTDIR) && go run ../scripts/dist/builddist.go -b $(BINNAME) -v $(VERSION)

.PHONY: clean
clean:
	rm -rf $(BINDIR) $(DISTDIR)
	go clean -cache

.PHONY: info
info: ## Get version info
	 @echo "Version:           ${VERSION}"
	 @echo "Git Tag:           ${GIT_TAG}"
	 @echo "Git Commit:        ${GIT_COMMIT}"
	 @echo "Git Tree State:    ${GIT_DIRTY}"

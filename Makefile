#
# Copyright 2019-2021 The sakuracloud_exporter Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
NAME     := sakuracloud_exporter
REVISION := $(shell git rev-parse --short HEAD)
SRCS     := $(shell find . -type f -name '*.go')
LDFLAGS  := -ldflags="-s -w -X \"main.Revision=$(REVISION)\" -extldflags -static"

PREFIX                  ?= $(shell pwd)/bin
BIN_DIR                 ?= $(shell pwd)/bin

AUTHOR          ?="The sakuracloud_exporter Authors"
COPYRIGHT_YEAR  ?="2019-2021"
COPYRIGHT_FILES ?=$$(find . -name "*.go" -print | grep -v "/vendor/")

GO     := GO111MODULE=on go
PKGS    = $(shell $(GO) list ./... | grep -v /vendor/)

default: lint test
all: lint test build
lint: fmt goimports
	@echo ">> running golangci-lint"
	golangci-lint run ./...

test:
	@echo ">> running tests"
	@$(GO) test -v $(PKGS)

fmt:
	@echo ">> formatting code"
	@$(GO) fmt $(PKGS)

goimports: fmt
	goimports -l -w $$(find . -type f -name '*.go' -not -path "./vendor/*")

run:
	@$(GO) run main.go

clean:
	rm -Rf $(BIN_DIR)/*

build: $(BIN_DIR)/$(NAME)

$(BIN_DIR)/$(NAME): $(SRCS)
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -a -tags netgo -installsuffix netgo -o $(BIN_DIR)/$(NAME)

.PHONY: tools
tools:
	GO111MODULE=off go get golang.org/x/tools/cmd/goimports
	GO111MODULE=off go get github.com/sacloud/addlicense
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/v1.38.0/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.38.0


.PHONY: set-license
set-license:
	@addlicense -c $(AUTHOR) -y $(COPYRIGHT_YEAR) $(COPYRIGHT_FILES)

.PHONY: all fmt build build-x test goimports clean lint

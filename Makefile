NAME     := sakuracloud_exporter
VERSION  := $(subst /,-,$(shell cat VERSION))
REVISION := $(shell git rev-parse --short HEAD)
SRCS     := $(shell find . -type f -name '*.go')
LDFLAGS  := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\" -extldflags -static"

PREFIX                  ?= $(shell pwd)/bin
BIN_DIR                 ?= $(shell pwd)/bin
DOCKER_IMAGE_NAME       ?= sacloud/sakuracloud_exporter
DOCKER_IMAGE_TAG        ?= $(subst /,-,$(shell cat VERSION))

AUTHOR          ?="The sakuracloud_exporter Authors"
COPYRIGHT_YEAR  ?="2019-2020"
COPYRIGHT_FILES ?=$$(find . -name "*.go" -print | grep -v "/vendor/")

GO     := GO111MODULE=on go
PKGS    = $(shell $(GO) list ./... | grep -v /vendor/)

all: fmt build test
lint: fmt vet style

test:
	@echo ">> running tests"
	@$(GO) test -v $(PKGS)

style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

fmt:
	@echo ">> formatting code"
	@$(GO) fmt $(PKGS)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(PKGS)

run:
	@$(GO) run main.go

clean:
	rm -Rf $(BIN_DIR)/*

build: $(BIN_DIR)/$(NAME)

$(BIN_DIR)/$(NAME): $(SRCS)
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -a -tags netgo -installsuffix netgo -o $(BIN_DIR)/$(NAME)

build-x:
	for os in darwin linux windows; do \
	    for arch in amd64 386; do \
	        GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 $(GO) build $(LDFLAGS) -a -tags netgo -installsuffix netgo -o $(BIN_DIR)/$(NAME); \
	        ( cd $(BIN_DIR); zip -r "$(NAME)_$$os-$$arch" $(NAME) ../LICENSE ../README.md ); \
	        rm -f $(BIN_DIR)/$(NAME); \
	    done; \
	done

docker:
	@echo ">> building docker image"
	@docker build -t "$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .

.PHONY: tools
tools:
	GO111MODULE=off go get github.com/sacloud/addlicense

.PHONY: set-license
set-license:
	@addlicense -c $(AUTHOR) -y $(COPYRIGHT_YEAR) $(COPYRIGHT_FILES)



.PHONY: all style fmt build build-x test vet docker clean lint

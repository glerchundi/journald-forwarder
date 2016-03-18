# MAINTAINER: Gorka Lerchundi Osa <glertxundi@gmail.com>

NAME       = journald
VERSION    = 0.1.0
PACKAGE    = github.com/glerchundi/journald-forwarder
GIT_REV    = `git rev-parse --verify HEAD`
BUILD_DATE = `date -u +"%Y-%m-%dT%H:%M:%SZ"`

forwarders := $(wildcard forwarder-*)

all: $(forwarders)

ifeq ($(BUILD),prod)
$(forwarders):
	@echo "Building production $(NAME)-$@..."
	ROOTPATH=$(shell pwd -P); mkdir -p $$ROOTPATH/bin; \
	GO15VENDOREXPERIMENT=1 \
	GOOS=linux GOARCH=amd64 \
	CGO_ENABLED=1 CGO_CPPFLAGS="-I $$ROOTPATH/core"  \
	go build \
		-a -x -tags netgo -installsuffix cgo -installsuffix netgo \
		-ldflags " \
		  -X $(PACKAGE)/core.Version=$(VERSION) \
		  -X $(PACKAGE)/core.GitRev=$(GIT_REV) \
		  -X $(PACKAGE)/core.BuildDate=$(BUILD_DATE) \
		" \
		-o $$ROOTPATH/bin/$(NAME)-$@-linux-amd64 \
		./$@
else
$(forwarders):
	@echo "Building $(NAME)-$@..."
	ROOTPATH=$(shell pwd -P); mkdir -p $$ROOTPATH/bin; \
	GO15VENDOREXPERIMENT=1 \
	CGO_ENABLED=1 CGO_CPPFLAGS="-I $$ROOTPATH/core"  \
	go build \
		-x \
		-ldflags " \
		  -X $(PACKAGE)/core.Version=$(VERSION) \
		  -X $(PACKAGE)/core.GitRev=$(GIT_REV) \
		  -X $(PACKAGE)/core.BuildDate=$(BUILD_DATE) \
		" \
		-o $$ROOTPATH/bin/$(NAME)-$@ \
		./$@
endif

test:
	@echo "Running tests..."
	@GO15VENDOREXPERIMENT=1 go test ./core
	@$(foreach forwarder,$(forwarders),GO15VENDOREXPERIMENT=1 go test ./$(forwarder);)

clean:
	rm -f bin/$(NAME)*

.PHONY: all $(forwarders) test clean
# Makefile for the Docker image quay.io/glerchundi/journald-forwarder
# MAINTAINER: Gorka Lerchundi Osa <glertxundi@gmail.com>
# If you update this image please bump the tag value before pushing.

#VERSION = 0.2.0
#PREFIX = quay.io/glerchundi
NAME = journald

forwarders := $(wildcard forwarder-*)

all: $(forwarders)

ifeq ($(BUILD),prod)
$(forwarders):
	@echo "Building static $(NAME)-$@..."
	ROOTPATH=$(shell pwd -P); mkdir -p $$ROOTPATH/bin; \
	GO15VENDOREXPERIMENT=1 \
	GOOS=linux GOARCH=amd64 \
	CGO_ENABLED=1 CGO_CPPFLAGS="-I $$ROOTPATH/core"  \
	go build \
		-a -x -tags netgo -installsuffix cgo -installsuffix netgo \
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
		-o $$ROOTPATH/bin/$(NAME)-$@ \
		./$@
endif

test:
	@echo "Running tests..."
	@GO15VENDOREXPERIMENT=1 go test ./core
	@$(foreach forwarder,$(forwarders),GO15VENDOREXPERIMENT=1 go test ./$(forwarder);)

#container: static
#	docker build -t $(PREFIX)/$(NAME):$(VERSION) .

#push: container
#	docker push $(PREFIX)/$(NAME):$(VERSION)

clean:
	rm -f bin/$(NAME)*

.PHONY: all $(forwarders) test clean
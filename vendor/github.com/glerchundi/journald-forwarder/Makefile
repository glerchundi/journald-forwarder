# Makefile for the Docker image quay.io/glerchundi/journald-forwarder
# MAINTAINER: Gorka Lerchundi Osa <glertxundi@gmail.com>
# If you update this image please bump the tag value before pushing.

#VERSION = 0.2.0
#PREFIX = quay.io/glerchundi
NAME = journald

FORWARDERS := $(wildcard forwarder-*)

all: $(FORWARDERS)

ifeq ($(BUILD),static)
$(FORWARDERS):
	@echo "Building static $(NAME)-$@..."
	ROOTPATH=$(shell pwd -P); mkdir -p $$ROOTPATH/bin; \
	GO15VENDOREXPERIMENT=1 \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build \
		-a -tags netgo -installsuffix cgo -ldflags '-extld ld -extldflags -static' -a -x \
		-o $$ROOTPATH/bin/$(NAME)-$@-linux-amd64 \
		./$@
else
$(FORWARDERS):
	@echo "Building $(NAME)-$@..."
	ROOTPATH=$(shell pwd -P); mkdir -p $$ROOTPATH/bin; \
	GO15VENDOREXPERIMENT=1 go build -o $$ROOTPATH/bin/$(NAME)-$@ ./$@
endif

test:
	@echo "Running tests..."
	GO15VENDOREXPERIMENT=1 go test

#container: static
#	docker build -t $(PREFIX)/$(NAME):$(VERSION) .

#push: container
#	docker push $(PREFIX)/$(NAME):$(VERSION)

clean:
	rm -f bin/$(NAME)*

.PHONY: all $(FORWARDERS) test clean

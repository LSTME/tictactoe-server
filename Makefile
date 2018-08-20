.PHONY: dockerfile-%, build-%, run-%, prod, clean

NAME ?= tictactoe-server
MAIN_PKG_PATH ?= .
ARCH ?= amd64
OS ?= linux
GOPROJECT_PATH ?= $(shell ./goproject-path.sh)

REGISTRY ?= lstme
IMAGE ?= $(REGISTRY)/$(NAME)-$(OS)-$(ARCH)
VERSION := $(shell git describe --tags --always --dirty)
ifeq ($(VERSION), '')
	VERSION = 'latest'
endif

BUILDIMAGE=golang:1.8.3-jessie
RUNIMAGE?=alpine

prod: dockerfile-prod build-prod

dockerfile-%: Dockerfile.%.in Makefile
	sed \
    -e 's|ARG_NAME|$(NAME)|g' \
    -e 's|ARG_GOARCH|$(ARCH)|g' \
    -e 's|ARG_GOOS|$(OS)|g' \
    -e 's|ARG_BUILDIMAGE|$(BUILDIMAGE)|g' \
    -e 's|ARG_RUNIMAGE|$(RUNIMAGE)|g' \
    -e 's|ARG_MAIN_PKG_PATH|$(MAIN_PKG_PATH)|g' \
    -e 's|ARG_GOPROJECT_PATH|$(GOPROJECT_PATH)|g' \
    Dockerfile.$*.in > Dockerfile

build-%: %
	docker build -t $(REGISTRY)/$(IMAGE):$(VERSION) .

run-%: build-%
	docker create -p 8987:32768 --name lstme_tictactoe-server $(REGISTRY)/$(IMAGE):$(VERSION) 

clean:
	rm -rf Dockerfile

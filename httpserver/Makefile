export tag=v1.1
REPO=lendea
IMAGE=httpserver_study
SHELL        := /bin/bash
GO           ?= go
GOFMT        ?= $(GO)fmt
TIME         := $(shell date "+%Y%m%d%H%M%S")
BRANCH       := $(shell git branch | grep \* | cut -d ' ' -f 2)
VERSION      := $(shell if [[ ${BRANCH} == "main" ]]; then \
                            git describe --tags `git rev-list --tags=v* --max-count=1`; \
                        else \
                            echo "dev-"`git rev-list HEAD --max-count=1 | cut -b 1-8`; \
                        fi)
BUILD        := ./build_tools/govvv build -version $(VERSION) -ldflags "-X main.BuildTime=$(TIME)"
ifeq ($(OS), mac)
BUILD        := ./build_tools/govvv-mac build -version $(VERSION) -ldflags "-X main.BuildTime=$(TIME)"
endif

STATICCHECK  := ./build_tools/staticcheck
ifeq ($(OS), mac)
STATICCHECK  := ./build_tools/staticcheck-mac
endif

pkgs          = ./...
TESTPKGS 	  = $(shell $(GO) list $(pkgs) | grep -v "tests")
ENV_NAME 	  = local

CERT_COMMON_NAME = localhost

.PHONY: all prod-build prod
all: prod
prebuild: format style staticcheck
prod: prod-build

.PHONY: format
format:
	@echo ">> formatting code"
	$(GO) fmt $(pkgs)

.PHONY: style
style:
	@echo ">> checking code style"
	! $(GOFMT) -w -d $$(find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

.PHONY: staticcheck
staticcheck:
	@echo ">> running staticcheck"
	$(STATICCHECK) $(pkgs)

prod-build: prebuild
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(BUILD) -o app .

prod-build-mac: prebuild
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(BUILD) -o app .


.PHONY: ut
#make ut OS=mac pkgs=./common
ut:
	@echo ">> running all tests"
	$(GO) test -gcflags=-l -cover -v -race $(TESTPKGS)

.PHONY: cert
cert:
	@echo ">> generating cert"
	cd cert && \
	openssl genrsa -out server.key 2048 && \
	openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650 -subj "/C=SH/ST=ShangHai/L=ShangHai/O=Lendea/OU=Study/CN=$(CERT_COMMON_NAME)/emailAddress=endealei@gmail.com"

.PHONY: release
release:
	echo "building httpserver container"
	docker build -t lendea/httpserver:${tag} .

.PHONY: push
push:
	echo "pushing lendea/httpserver"
	docker push lendea/httpserver:${tag}
# The MIT License (MIT)

# Copyright (c) 2014 traetox

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:

# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

APP = speedtest

SHELL = /bin/bash

DIR = $(shell pwd)

GO = go
GLIDE = glide

# GOX = gox -os="linux darwin windows freebsd openbsd netbsd"
GOX = gox -os="linux darwin windows"

NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

MAIN = github.com/traetox/speedtest
SRCS = $(shell git ls-files '*.go' | grep -v '^vendor/')
PKGS = $(shell glide novendor)
EXE = speedtest

VERSION=$(shell \
        grep "const Version" version/version.go \
        |awk -F'=' '{print $$2}' \
        |sed -e "s/[^0-9.]//g" \
	|sed -e "s/ //g")

PACKAGE=$(APP)-$(VERSION)
ARCHIVE=$(PACKAGE).tar

all: help

help:
	@echo -e "$(OK_COLOR)==== $(APP) [$(VERSION)] ====$(NO_COLOR)"
	@echo -e "$(WARN_COLOR)init$(NO_COLOR)      :  Install requirements"
	@echo -e "$(WARN_COLOR)build$(NO_COLOR)     :  Make all binaries"
	@echo -e "$(WARN_COLOR)test$(NO_COLOR)      :  Launch unit tests"
	@echo -e "$(WARN_COLOR)lint$(NO_COLOR)      :  Launch golint"
	@echo -e "$(WARN_COLOR)vet$(NO_COLOR)       :  Launch go vet"
	@echo -e "$(WARN_COLOR)coverage$(NO_COLOR)  :  Launch code coverage"
	@echo -e "$(WARN_COLOR)clean$(NO_COLOR)     :  Cleanup"
	@echo -e "$(WARN_COLOR)binaries$(NO_COLOR)  :  Make binaries"

clean:
	@echo -e "$(OK_COLOR)[$(APP)] Cleanup$(NO_COLOR)"
	@rm -fr $(EXE) $(APP)-*.tar.gz

.PHONY: init
init:
	@echo -e "$(OK_COLOR)[$(APP)] Install requirements$(NO_COLOR)"
	@go get -u github.com/golang/glog
	@go get -u github.com/Masterminds/glide
	@go get -u github.com/golang/lint/golint
	@go get -u github.com/kisielk/errcheck
	@go get -u golang.org/x/tools/cmd/oracle
	@go get -u github.com/mitchellh/gox

.PHONY: deps
deps:
	@echo -e "$(OK_COLOR)[$(APP)] Update dependencies$(NO_COLOR)"
	@glide up -u -s

.PHONY: build
build:
	@echo -e "$(OK_COLOR)[$(APP)] Build $(NO_COLOR)"
	@$(GO) build .

.PHONY: test
test:
	@echo -e "$(OK_COLOR)[$(APP)] Launch unit tests $(NO_COLOR)"
	@$(GO) test -v $$(glide nv)

.PHONY: lint
lint:
	@$(foreach file,$(SRCS),golint $(file) || exit;)

.PHONY: vet
vet:
	@$(foreach file,$(SRCS),$(GO) vet $(file) || exit;)

.PHONY: errcheck
errcheck:
	@echo -e "$(OK_COLOR)[$(APP)] Go Errcheck $(NO_COLOR)"
	@$(foreach pkg,$(PKGS),errcheck $(pkg) $(glide novendor) || exit;)

.PHONY: coverage
coverage:
	@$(foreach pkg,$(PKGS),$(GO) test -cover $(pkg) $(glide novendor) || exit;)

.PHONY: binaries
binaries:
	@echo -e "$(OK_COLOR)[$(APP)] Create binaries $(NO_COLOR)"
	$(GOX) github.com/traetox/speedtest

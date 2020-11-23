PWD := $(shell pwd)
VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)
PROJECTNAME := $(shell basename $(PWD))
USERNAME := $(shell git config user.name)
GOOS := windows
GOARCH := amd64
TAG := $(VERSION)_$(GOOS)_$(GOARCH)
PLATFORMS=darwin linux windows
ARCHITECTURES=386 amd64

# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-w -s -X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

.PHONY: build

build: buildwithoutdebug pack

buildfordebug:
	@go build -o build/$(PROJECTNAME)_$(TAG).exe -v ./

buildwithoutdebug:
	@go build $(LDFLAGS) -o build/$(PROJECTNAME)_$(TAG).exe -v ./

buildwodebug_linux:
	@set GOOS=linux&& go build $(LDFLAGS) -o build/$(PROJECTNAME)_$(TAG) -v ./

buildwithoutdebug_linux:
	@set GOARCH=$(GOARCH)&&set GOOS=$(GOOS)
	@go build $(LDFLAGS) -o build/$(PROJECTNAME)_$(VERSION)_$(GOOS)_$(GOARCH) -v ./

prebuild_all:
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); go build -v $(LDFLAGS) -o build/$(PROJECTNAME)_$(VERSION)_$(GOOS)_$(GOARCH))))

build_all: prebuild_all pack

run: build
	build/$(PROJECTNAME)_$(TAG).exe
	
.DUFAULT_GOAL := build

pack:
	upx --ultra-brute build/$(PROJECTNAME)*

mod_init:
	go mod init github.com/$(USERNAME)/$(PROJECTNAME)

mod:
	go mod tidy
	go mod download
	go mod vendor

install:
	go install ${LDFLAGS}

# Remove only what we've created
clean:
	find ${PWD} -name 'build/${PROJECTNAME}[-?][a-zA-Z0-9]*[-?][a-zA-Z0-9]*' -delete
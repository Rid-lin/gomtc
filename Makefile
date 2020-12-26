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
UPX := $(shell /mnt/c/apps/upx.exe)

# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-w -s -X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# Check for required command tools to build or stop immediately
EXECUTABLES = git go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

.PHONY: build

build: buildwithoutdebug pack

.PHONY: buildfordebug
buildfordebug:
	@go build -o build/$(PROJECTNAME)_$(VERSION)_$(GOOS)_$(GOARCH).exe -v ./

.PHONY: buildwithoutdebug
buildwithoutdebug:
	@go build $(LDFLAGS) -o build/$(PROJECTNAME)_$(VERSION)_$(GOOS)_$(GOARCH).exe -v ./

.PHONY: buildwodebug_linux
buildwodebug_linux:
	$(shell export GOOS=linux; go build $(LDFLAGS) -o build/$(PROJECTNAME)_$(VERSION)_$(GOOS)_$(GOARCH) -v ./)

.PHONY: buildwithoutdebug_linux
buildwithoutdebug_linux:
	@set GOARCH=$(GOARCH)&&set GOOS=$(GOOS)
	@go build $(LDFLAGS) -o build/$(PROJECTNAME)_$(VERSION)_$(GOOS)_$(GOARCH) -v ./

.PHONY: prebuild_all
prebuild_all:
	$(foreach GOOS, $(PLATFORMS),\
	$(foreach GOARCH, $(ARCHITECTURES), $(shell export GOOS=$(GOOS); export GOARCH=$(GOARCH); go build -v $(LDFLAGS) -o build/$(PROJECTNAME)_$(VERSION)_$(GOOS)_$(GOARCH))))


.PHONY: build_all
build_all: prebuild_all pack


.PHONY: run
run: build
	build/$(PROJECTNAME)_$(VERSION)_$(GOOS)_$(GOARCH).exe
	
.DUFAULT_GOAL := prebuild_all


.PHONY: pack
pack:
	$(UPX) --ultra-brute build/$(PROJECTNAME)*

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
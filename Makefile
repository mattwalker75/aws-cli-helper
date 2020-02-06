
BUILD := $(shell git rev-parse --short HEAD)
PROJECTNAME := $(shell basename "$(PWD)")

# Go related variables.
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOBIN := $(GOBASE)/bin

# Use linker flags to provide version/build settings ( -s and -w helps make the binary smaller but can affect performance )
LDFLAGS=-ldflags "-s -w -X=main.Build=$(BUILD)"

# Make is verbose in Linux. Make it silent.
MAKEFLAGS += --silent

## reset: Delete all directories and files not part of the repo.  Perform a reset on the repo for commits
reset:
	@-$(MAKE) go-clean
	@-rm -Rf $(GOBIN)
	@-rm -Rf $(GOBASE)/vendor

## clean: Cleans up build files.
clean:
	@-rm $(GOBIN)/$(PROJECTNAME) 2> /dev/null
	@-$(MAKE) go-clean

## install: Install missing dependencies needed by your programs. 
install: go-get

## compile: Runs "install" and then compiles the programs and put them in the bin directory
compile: go-compile

go-compile: go-get go-build

go-build:
	@echo "  >  Building binary..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build $(LDFLAGS) -o $(GOBIN)/ListEC2 ListEC2.go
	@echo " Files:"
	@echo "   - $(GOBIN)/ListEC2"
	@echo ""

go-get:
	@echo "  >  Checking if there is any missing dependencies..."
	#@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get $(get)
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get -d

go-clean:
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean


.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo


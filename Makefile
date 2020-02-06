
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

# This "install" option gets ran by default if no parameters are passed to the "make" command because it is the first one
# in the list.
## install: Runs "dep_install" and then compiles the programs and put them in the bin directory
install: go-compile

## dep_install: Install missing dependencies needed by your programs. 
dep_install: go-get

## reset: Delete all directories and files not part of the repo.  Perform a reset on the repo for commits
reset:
	@-$(MAKE) go-clean
	@echo "  >  Deleting the bin directory..."
	@-rm -Rf $(GOBIN)
	@echo "  >  Deleting the vendor directory..."
	@-rm -Rf $(GOBASE)/vendor

## clean: Cleans up build files.
clean:
	@-rm $(GOBIN)/$(PROJECTNAME) 2> /dev/null
	@-$(MAKE) go-clean

go-compile: go-get go-build

go-build:
	@echo "  >  Building binaries..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build $(LDFLAGS) -o $(GOBIN)/ListEC2 ListEC2.go
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build $(LDFLAGS) -o $(GOBIN)/ListENIs ListENIs.go
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build $(LDFLAGS) -o $(GOBIN)/VPCFlowLogData VPCFlowLogData.go
	@echo " Files:"
	@echo "   - $(GOBIN)/ListEC2"
	@echo "   - $(GOBIN)/ListENIs"
	@echo "   - $(GOBIN)/VPCFLowLogData"
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


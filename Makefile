#!/usr/bin/make -f

PKGS := $(shell go list ./cmd/...)
BINS =  $(shell basename $(PKGS))
COVERAGE_REPORT_FILENAME ?= coverage.out
BUILDDIR ?= $(CURDIR)/build

ifeq (,$(findstring nostrip,$(BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

ifneq (,$(ldflags))
  BUILD_FLAGS += -ldflags '$(ldflags)'
endif

# check for nostrip option
ifeq (,$(findstring nostrip,$(BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

# Check for debug option
ifeq (debug,$(findstring debug,$(BUILD_OPTIONS)))
  BUILD_FLAGS += -gcflags "all=-N -l"
endif

# Check for the verbose option
ifdef verbose
VERBOSE = -v
endif

all: build check

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/

$(BUILD_TARGETS): generate $(BUILDDIR)/
	go $@ $(VERBOSE) -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $@

check: $(COVERAGE_REPORT_FILENAME)

$(COVERAGE_REPORT_FILENAME): generate
	go test $(VERBOSE) -mod=readonly -race -cover -covermode=atomic -coverprofile=$@ ./...

go.sum: go.mod
	echo "Ensure dependencies have not been modified ..." >&2
	go mod verify
	go mod tidy
	touch $@

generate: generate-stamp
generate-stamp: go.sum
	go generate ./...
	touch $@

distclean: clean
	rm -rf dist/

clean:
	rm -rf $(BUILDDIR)
	rm -f \
	   $(COVERAGE_REPORT_FILENAME) \
	   generate-stamp

list:
	@echo $(BINS) | tr ' ' '\n'

.PHONY: all clean check distclean build list

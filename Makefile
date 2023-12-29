#!/usr/bin/make -f

PKGS := $(shell go list ./cmd/...)
BINS =  $(shell basename $(PKGS))
COVERAGE_REPORT_FILENAME ?= coverage.out
BUILDDIR ?= $(CURDIR)/build
CODESIGN_IDENTIY ?= none

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

all: version.txt build check

BUILD_TARGETS := build install

build: version.txt
build: BUILD_ARGS=-o $(BUILDDIR)/

$(BUILD_TARGETS): $(BUILDDIR)/
	go $@ $(VERBOSE) -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $@

check: $(COVERAGE_REPORT_FILENAME)

$(COVERAGE_REPORT_FILENAME): version.txt
	go test $(VERBOSE) -mod=readonly -race -cover -covermode=atomic -coverprofile=$@ ./...

deps:
	echo "Ensure dependencies have not been modified ..." >&2
	go mod verify
	go generate ./...
	go mod tidy

distclean: clean
	rm -rf dist/
	rm -rf unixtools.dmg

clean:
	rm -rf $(BUILDDIR)
	rm -f \
	   $(COVERAGE_REPORT_FILENAME) \
	   version.txt

version.txt: deps
	cp -f version/version.txt version.txt

list:
	@echo $(BINS) | tr ' ' '\n'

macos-codesign: build
	codesign --verbose -s $(CODESIGN_IDENTITY) --options=runtime $(BUILDDIR)/*

unixtools.pkg: version.txt macos-codesign
	pkgbuild --identifier io.asscrypto.unixtools \
		--install-location ./Library/ --root $(BUILDDIR) $@

unixtools.dmg: version.txt macos-codesign
	VERSION=$(shell cat version.txt); \
	mkdir -p dist/unixtools-$${VERSION}/bin ; \
	cp -a $(BUILDDIR)/* dist/unixtools-$${VERSION}/bin/ ; \
	chmod 0755 dist/unixtools-$${VERSION}/bin/* ; \
	create-dmg --volname unixtools --codesign $(CODESIGN_IDENTITY) \
		--sandbox-safe --no-internet-enable \
		$@ dist/unixtools-$${VERSION}

.PHONY: all clean check distclean build list macos-codesign deps

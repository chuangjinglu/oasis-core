SHELL := /bin/bash

# Check if we're running in an interactive terminal.
ISATTY := $(shell [ -t 0 ] && echo 1)

ifdef ISATTY
	# Running in interactive terminal, OK to use colors!
	MAGENTA := \e[35;1m
	CYAN := \e[36;1m
	OFF := \e[0m

	# Built-in echo doesn't support '-e'.
	ECHO = /bin/echo -e
else
	# Don't use colors if not running interactively.
	MAGENTA := ""
	CYAN := ""
	OFF := ""

	# OK to use built-in echo.
	ECHO := echo
endif

# Try to determine Oasis Core's version from git.
LATEST_TAG := $(shell git describe --tags --match 'v*' --abbrev=0 2>/dev/null)
VERSION := $(subst v,,$(LATEST_TAG))
IS_TAG := $(shell git describe --tags --match 'v*' --exact-match 2>/dev/null && echo YES || echo NO)
ifeq ($(and $(LATEST_TAG),$(IS_TAG)),NO)
	# The current commit is not exactly a tag, append commit and dirty info to
	# the version.
	VERSION := $(VERSION)-git$(shell git describe --always --match '' --dirty=+dirty 2>/dev/null)
endif

# Go binary to use for all Go commands.
OASIS_GO ?= go

# Go command prefix to use in all Go commands.
GO := env -u GOPATH $(OASIS_GO)

# NOTE: The -trimpath flag strips all host dependent filesystem paths from
# binaries which is required for deterministic builds.
GOFLAGS ?= -trimpath -v

# Add Oasis Core's version as a linker string value definition.
ifneq ($(VERSION),)
	GOLDFLAGS += "-X github.com/oasislabs/oasis-core/go/common/version.SoftwareVersion=$(VERSION)"
endif

# Go build command to use by default.
GO_BUILD_CMD := env -u GOPATH $(OASIS_GO) build $(GOFLAGS)

# Path to the Urkel interoperability test helpers binary in go/.
GO_TEST_HELPER_URKEL_PATH := storage/mkvs/urkel/interop/urkel-test-helpers

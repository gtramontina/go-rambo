SHELL := /usr/bin/env bash -eu -o pipefail
CPUS ?= $(shell (nproc --all || sysctl -n hw.ncpu) 2>/dev/null || echo 1)
MAKEFLAGS += --warn-undefined-variables --output-sync=line --jobs $(CPUS)
.DEFAULT_GOAL := help
.SECONDEXPANSION:
.DELETE_ON_ERROR:

.git/hooks.log:
	git config core.hooksPath .githooks
	git config --get core.hooksPath > $@
pre_reqs += .git/hooks.log
to_trash += .git/hooks.log

# ---

source_files = $(shell {\
	git ls-files -- '*.go'; \
	git ls-files --others --exclude-standard -- '*.go'; \
})

#: Installs all pre-requisites.
install: $(pre_reqs)
.PHONY: install

#: Runs all tests.
test: | $(pre_reqs)
	go test -cover -race -count=1 -test.shuffle=on ./...
.PHONY: test

#: Runs all benchmarks.
bench: | $(pre_reqs)
	go test -bench=. -benchmem ./...
.PHONY: bench

#: Checks if there are any lint errors.
lint: | $(pre_reqs)
	test -z $$(gofmt -l $(source_files))
.PHONY: lint

#: Fixes lint errors.
lint.fix: | $(pre_reqs)
	gofmt -w -s $(source_files)
.PHONY: lint.fix

#: Removes all generated files.
clobber:
	rm -rf $(to_trash)
.PHONY: clobber

#: Fixes lint errors and runs tests. This runs before all commits.
pre-commit: | $(pre_reqs)
	MAKEFLAGS= $(MAKE) lint.fix
	go test -cover ./...
.PHONY: pre-commit

# ---

#: Prints the list of PHONY targets. This is the default target.
help:
ifndef help
	@echo -e "\nAvailable phony targets:\n"
	@help=true MAKEFLAGS= $(MAKE) -rpn \
	| sed -rn "s/^\.PHONY: (.*)/\1/p" | tr " " "\n" \
	| sort -u \
	| sed -re "s/^($(.DEFAULT_GOAL))$$/\1 $$(tput setaf 2)(default)$$(tput sgr0)/" \
	| sed -e "s/^/  $$(tput setaf 8)make$$(tput sgr0) /"
endif
.PHONY: help

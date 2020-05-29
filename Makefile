SHELL:=/bin/bash

changelog_args=-o CHANGELOG.md -p '^v'

lint:
	golangci-lint run --concurrency 4 --print-issued-lines=false --exclude-use-default=false --enable=golint --enable=goimports  --enable=unconvert --enable=unparam

changelog:
ifdef version
	$(eval changelog_args=--next-tag $(version) $(changelog_args))
endif
	git-chglog $(changelog_args)

check-gotest:
ifeq (, $(shell which richgo))
	$(warning "richgo is not installed, falling back to plain go test")
	$(eval TEST_BIN=go test)
else
	$(eval TEST_BIN=richgo test)
endif
	$(eval test_command=$(TEST_BIN) ./... -v --cover)

test-only: check-gotest
	$(test_command)

test: lint test-only
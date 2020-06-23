# Task runner

.PHONY: help build

.DEFAULT_GOAL := help

SHELL := /bin/bash

# App version is the tag attached to this commit, if there is one. If there is not, APP_VERSION is empty. If there are
# multiple, they are concatenated using the " " (space) character.
#
# See http://stackoverflow.com/questions/4545370/how-to-list-all-tags-pointing-to-a-specific-commit-in-git
# See http://stackoverflow.com/questions/1404796/how-to-get-the-latest-tag-name-in-current-branch-in-git
APP_VERSION  := $(shell git show-ref --dereference --tags | grep ^`git rev-parse HEAD` | sed -e 's,.* refs/tags/,,' -e 's/\^{}//')
GIT_HASH     := $(shell git rev-parse --short HEAD)

ANSI_TITLE        := '\e[1;32m'
ANSI_CMD          := '\e[0;32m'
ANSI_TITLE        := '\e[0;33m'
ANSI_SUBTITLE     := '\e[0;37m'
ANSI_WARNING      := '\e[1;31m'
ANSI_OFF          := '\e[0m'

PATH_DOCS                := $(shell pwd)/docs
PATH_BUILD_CONFIGURATION := $(shell pwd)/build

TIMESTAMP := $(shell date "+%s")

help: ## Show this menu
	@echo -e $(ANSI_TITLE)library$(ANSI_OFF)$(ANSI_SUBTITLE)" - FILL_THIS_IN_WITH_SOMETHING_LOGICAL"$(ANSI_OFF)
	@echo -e "\nUsage: $ make \$${COMMAND} \n"
	@echo -e "Variables use the \$${VARIABLE} syntax, and are supplied as environment variables before the command. For example, \n"
	@echo -e "  \$$ VARIABLE="foo" make help\n"
	@echo -e $(ANSI_TITLE)Commands:$(ANSI_OFF)
	@grep -E '^[a-z.A-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[32m%-30s\033[0m %s\n", $$1, $$2}'

clean: ## Cleans up build artifacts
	rm -rf dist

bin: clean ## Builds the binaries
	mkdir -p dist/usr/bin
	cd src && \
		go build -o ../dist/usr/bin/library

docker.build: ## Builds the container
	docker build \
		--tag gcr.io/littleman-co/library:latest \
		--tag gcr.io/littleman-co/library:$(GIT_HASH) \
		.

docker.push: docker.build ## Pushes the container
	docker push gcr.io/littleman-co/library:latest
	docker push gcr.io/littleman-co/library:$(GIT_HASH)
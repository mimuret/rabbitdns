# Shortcut targets
default: build

## Build binary
all: build

## Run the tests
test: ut

build: gen-files

# Define some constants
#######################
GO_BUILD_VER     ?= v12.0.0
RABBITDNS_BUILD  ?= rabbitdns/go-build:$(GO_BUILD_VER)
PACKAGE_NAME     ?= rabbitdns/rabbitdns
LOCAL_USER_ID    ?= $(shell id -u $$USER)
BINDIR           ?= bin
RABBITDNS_PKG     = github.com/rabbitdns/rabbitdns
TOP_SRC_DIR       = lib
MY_UID           := $(shell id -u)

DOCKER_GO_BUILD := mkdir -p .go-pkg-cache && \
                   docker run --rm \
                              --net=host \
                              $(EXTRA_DOCKER_ARGS) \
                              -e LOCAL_USER_ID=$(LOCAL_USER_ID) \
                              -v $(CURDIR):/go/src/github.com/$(PACKAGE_NAME):rw \
                              -w /go/src/github.com/$(PACKAGE_NAME) \
                              $(RABBITDNS_BUILD)

# Create a list of files upon which the generated file depends, skip the generated file itself
APIS_SRCS := $(filter-out ./pkg/apis/v1alpha/zz_generated.deepcopy.go, $(wildcard ./pkg/apis/v1alpha/*.go))

.PHONY: clean
## Removes all .coverprofile files, the vendor dir, and .go-pkg-cache
clean:
	find . -name '*.coverprofile' -type f -delete
	rm -rf vendor .go-pkg-cache
	rm -rf $(BINDIR)
	rm -rf checkouts

.PHONY: clean-gen-files
## Convenience target for devs to remove generated files and related utilities
clean-gen-files:
	rm -f $(BINDIR)/deepcopy-gen
	rm -f $(BINDIR)/client-gen
	rm -f $(BINDIR)/lister-gen
	rm -f ${RABBITDNS_PKG}/pkg/listers
	rm -f ${RABBITDNS_PKG}/pkg/clientset

vendor: Gopkg.lock
	$(DOCKER_GO_BUILD) dep ensure

.PHONY: gen-files
gen-files: ./pkg/apis/v1alpha/zz_generated.deepcopy.go \
					 ./pkg/clientset \
					 ./pkg/listers


./pkg/apis/v1alpha/zz_generated.deepcopy.go: ${APIS_SRCS} $(BINDIR)/deepcopy-gen
	$(DOCKER_GO_BUILD) \
		sh -c 'deepcopy-gen \
			--v 1 --logtostderr \
			--go-header-file "./docs/boilerplate.go.txt" \
			--input-dirs "$(RABBITDNS_PKG)/pkg/apis/v1alpha" \
			--bounding-dirs "github.com/rabbitdns/rabbitdns" \
			--output-file-base zz_generated.deepcopy'


./pkg/clientset: ./pkg/apis/v1alpha/zz_generated.deepcopy.go
	$(DOCKER_GO_BUILD) \
		sh -c client-gen \
			--clientset-name "v1" \ 
			--input-base "" \
			--input "$(RABBITDNS_PKG)/pkg/apis/v1alpha" \
			--output-package "${RABBITDNS_PKG}/pkg/clientset/v1alpha"


./pkg/listers: ./pkg/apis/v1alpha/zz_generated.deepcopy.go
	$(DOCKER_GO_BUILD) \
		sh -c lister-gen \
			--input-dirs "$(RABBITDNS_PKG)/pkg/apis/v1alpha" \
			--output-package "${RABBITDNS_PKG}/pkg/listers"

./pkg/listers: ${APIS_SRCS}
	$(DOCKER_GO_BUILD) \
		sh -c informer-gen \
			--input-dirs "$(RABBITDNS_PKG)/pkg/apis/v1alpha" \
			--versioned-clientset-package "${RABBITDNS_PKG}/pkg/clientset/v1alpha" \
			--listers-package "${PACKAGE_NAME}/pkg/listers" \
			--output-package "${PACKAGE_NAME}/pkg/informers" \

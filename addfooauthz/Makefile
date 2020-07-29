PROJECT_NAME = gokit-istio-security
BINARY_PREFIX = ${PROJECT_NAME}
PACKAGE_PREFIX = github.com/cage1016
IMAGE_PREFIX = cage1016/${BINARY_PREFIX}
BUILD_DIR = build
SERVICES = add authz foo
DOCKERS = $(addprefix dev_docker_,$(SERVICES))
STAGES = dev debug prod
CGO_ENABLED ?= 0
GOOS ?= linux
COMMIT_HASH = $(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP = $(shell date +%Y-%m-%dT%T%z)
BUILD_VERSION ?= 1.0.0
BUILD_VERSION_INFO = service.CommitHash=$(COMMIT_HASH) service.BuildTimeStamp=$(BUILD_TIMESTAMP) service.Version=${BUILD_VERSION}
PACKAGE_VERSION = $(addprefix ${PACKAGE_PREFIX}/${PROJECT_NAME}/internal/app/_/,$(BUILD_VERSION_INFO))
PACKAGE_VERSION_LDFLAGS = $(foreach n, $(PACKAGE_VERSION), $(foreach m, $(SERVICES), $(subst _,$(m),-X $(n))))
GOLDFLAGS = -s -w
DEBUG_GOGCFLAGS = -gcflags='all=-N -l' -ldflags "$(GOLDFLAGS)"
GOGCFLAGS = -ldflags "$(GOLDFLAGS)"
SHELL  := env BUILD_TAGS=$(BUILD_TAGS) $(SHELL)
BUILD_TAGS ?= "dev"

define compile_service
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) go build -tags ${BUILD_TAGS} $(2) -ldflags "$(PACKAGE_VERSION_LDFLAGS)" -o ${BUILD_DIR}/${BINARY_PREFIX}-$(1) cmd/$(1)/main.go
endef

define make_docker
	docker build --build-arg COMMIT_HASH=$(COMMIT_HASH) --build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP) --build-arg VERSION=$(BUILD_VERSION) --build-arg BINARY=${BINARY_PREFIX}-$(1) --tag=${IMAGE_PREFIX}-$(1):${BUILD_TAGS} -f deployments/docker/$(2) ./build
endef

all: help

.PHONY: all $(SERVICES) dev_dockers debug_dockers cleanbuild_dockers test help integration

cleandocker:
	# Remove retailbase containers
	docker ps -f name=${IMAGE_PREFIX}-* -aq | xargs docker rm -f
	# Remove old retailbase images
	docker images -q ${IMAGE_PREFIX}-* | xargs docker rmi -f

# Clean ghost docker images
cleanghost:
	# Remove exited containers
	docker ps -f status=dead -f status=exited -aq | xargs docker rm -v
	# Remove unused images
	docker images -f dangling=true -q | xargs docker rmi -f
	# Remove unused volumes
	docker volume ls -f dangling=true -q | xargs docker volume rm -f

install:
	cp ${BUILD_DIR}/* $(GOBIN)

## test: run unit test
test:
	# DEBUG=true bash -c "go test -v github.com/cage1016/gokit-istio-security/<package-name> -run ..."
	go test -v -race -tags=uint ./...

## integration: run integration test
integration:
	go test -v -race -tags=integration ./...

PD_SOURCES:=$(shell find ./pb -type d)
proto:
	@for var in $(PD_SOURCES); do \
		if [ -f "$$var/compile.sh" ]; then \
			cd $$var && ./compile.sh; \
			echo "complie $$var/$$(basename $$var).proto"; \
			cd $(PWD); \
		fi \
	done

# Regenerates OPA data from rego files
HAVE_GO_BINDATA := $(shell command -v go-bindata 2> /dev/null)
generate:
ifndef HAVE_GO_BINDATA
	@echo "requires 'go-bindata' (go get -u github.com/kevinburke/go-bindata/go-bindata)"
	@exit 1 # fail
else
	go generate ./...
endif

$(SERVICES):
	$(call compile_service,$(@),${GOGCFLAGS})

# docker build --cache-from=docker.io/cage1016/gokit-istio-security-cleanbuild-base:latest -t docker.io/cage1016/gokit-istio-security-cleanbuild-base:latest --file=deployments/docker/Dockerfile.base.cleanbuild .
$(DOCKERS):
	@echo BUILD_TAGS=${BUILD_TAGS}
	@echo BUILD_VERSION=${BUILD_VERSION}
	@echo BUILD_TIMESTAMP=${BUILD_TIMESTAMP}
	@echo COMMIT_HASH=${COMMIT_HASH}

	@if [ "$(filter $(@:dev_docker_%=%), $(SERVICES))" != "" ]; then\
		$(call compile_service,$(subst dev_docker_,,$(@)),${GOGCFLAGS});\
		$(call make_docker,$(subst dev_docker_,,$(@)),Dockerfile);\
		if [ "$(PUSH_IMAGE)" == "true" ]; then \
			docker push ${IMAGE_PREFIX}-$(subst dev_docker_,,$(@)); \
		fi \
	fi

## services: build all services
services: $(SERVICES)

## dev_dockers: quick build all
dev_dockers: $(DOCKERS)

## help: Prints this help message
helps=$(foreach s, $(SERVICES), "$(s): build $(s) service")
helps+=$(foreach s, $(DOCKERS), "$(s): build $(subst dev_docker_,,$(s)) docker image")
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
	@echo ""
	@for var in $(helps); do \
		echo $$var; \
	done | column -t -s ':' |  sed -e 's/^/  /'
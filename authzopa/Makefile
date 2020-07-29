PROJECT_NAME = adapter
BINARY_PREFIX = ${PROJECT_NAME}
PACKAGE_PREFIX = github.com/cage1016
IMAGE_PREFIX = cage1016/${BINARY_PREFIX}
BUILD_DIR = build
SERVICES = authzopa
DOCKERS = $(addprefix dev_docker_,$(SERVICES))
CGO_ENABLED ?= 0
GOOS ?= linux
GOLDFLAGS = -s -w
GOGCFLAGS = -ldflags "$(GOLDFLAGS)"
BUILD_TAGS ?= "dev"

define compile_service
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) go build -tags ${BUILD_TAGS} $(2)  -o ${BUILD_DIR}/${BINARY_PREFIX}-$(1) cmd/main.go
endef

define make_docker
	docker build --build-arg BINARY=${BINARY_PREFIX}-$(1) --tag=${IMAGE_PREFIX}-$(1):${BUILD_TAGS} -f deployments/docker/$(2) ./build
endef

all: help

.PHONY: all $(SERVICES) help

cleandocker:
	# Remove retailbase containers
	docker ps -f name=${IMAGE_PREFIX}-* -aq | xargs docker rm
	# Remove old retailbase images
	docker images -q ${IMAGE_PREFIX}-* | xargs docker rmi

# Clean ghost docker images
cleanghost:
	# Remove exited containers
	docker ps -f status=dead -f status=exited -aq | xargs docker rm -v
	# Remove unused images
	docker images -f dangling=true -q | xargs docker rmi
	# Remove unused volumes
	docker volume ls -f dangling=true -q | xargs docker volume rm

PD_SOURCES:=$(shell find ./pb -type d)
proto:
	@for var in $(PD_SOURCES); do \
		if [ -f "$$var/compile.sh" ]; then \
			cd $$var && ./compile.sh; \
			echo "complie $$var/$$(basename $$var).proto"; \
			cd $(PWD); \
		fi \
	done

## generate: go generate proto
generate:
	REPO_ROOT=${GOPATH}/src/istio.io/istio ISTIO_OUT=${GOPATH}/src/istio.io/istio LOCAL_OUT=${GOPATH}/src/istio.io/istio go generate ./...

$(SERVICES):
	$(call compile_service,$(@),${GOGCFLAGS})

$(DOCKERS):
	@if [ "$(filter $(@:dev_docker_%=%), $(SERVICES))" != "" ]; then\
		$(call compile_service,$(subst dev_docker_,,$(@)),${GOGCFLAGS});\
		$(call make_docker,$(subst dev_docker_,,$(@)),Dockerfile);\
	fi

## run: run authzopa 
run:
	go run cmd/main.go 44225

## mixs_server: run mixs server
mixs_server:
	${GOPATH}/src/istio.io/istio/out/darwin_amd64/mixs server --configStoreURL=fs://${GOPATH}/src/istio.io/istio/mixer/adapter/authzopa/testdata --log_output_level=attributes:debug

method?=POST
path?=/api/v1/add/sum
## mixc_check_ok: run mixc check 200 ok
mixc_check_ok:
	${GOPATH}/src/istio.io/istio/out/darwin_amd64/mixc check -s request.path="${path}",request.method="${method}" --stringmap_attributes "request.headers=x-jwt-playload:$(shell cat owner.json | tr -d "\n" | tr -d " \t\n\r" | base64  | tr -d '=' | tr '/+' '_-')"

## mixc_check_permission_deny: run mixc check 403 permission denied
mixc_check_permission_deny:
	${GOPATH}/src/istio.io/istio/out/darwin_amd64/mixc check -s request.path="${path}",request.method="${method}" --stringmap_attributes "request.headers=x-jwt-playload:$(shell cat viewer.json | tr -d "\n" | tr -d " \t\n\r" | base64  | tr -d '=' | tr '/+' '_-')"

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
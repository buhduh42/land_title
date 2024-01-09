#TODO need to think about these build-args now, defaults?, consistency?
#TODO, yup....
ROOT_DIR ?= $(realpath ${DIR}/../..)
SCRIPTS_DIR ?= ${ROOT_DIR}/scripts
SRC_DIR := ${DIR}/src

GO_VENDOR := ${SRC_DIR}/vendor

BUILD_DIR := ${DIR}/build
BUILD_DIRS := ${BUILD_DIR}
BUILD := ${BUILD_DIR}/build
TEST := ${BUILD_DIR}/test

DOCKERFILE ?= ${ROOT_DIR}/resources/Dockerfile

SERVICE ?= $(shell basename ${DIR})
#SRC := $(filter-out ${GO_LIBS}/%, $(shell find ${SRC_DIR} -type f -name '*.go')) ${DOCKERFILE}
SRC := $(filter-out ${GO_VENDOR}/%,$(shell find ${SRC_DIR} -type f -name '*.go')) ${DOCKERFILE}
GO_VERSION ?= $(shell ${SCRIPTS_DIR}/scrape_go_version.sh ${SRC_DIR}/go.mod)
GO_LIBS := ${GO_VENDOR}/landtitle

ENTRYPOINT ?= ${BUILD_DIR}/entrypoint
SERVICE_IMAGE ?= alpine:latest
CREATE_SERVICE := ${BUILD_DIR}/${SERVICE}_image

define GLOBAL_BUILD_ARGS
--build-arg SERVICE_IMAGE=${SERVICE_IMAGE} \
--build-arg GO_VERSION=${GO_VERSION}       \
--build-arg SERVICE=${SERVICE}
endef

.PHONY: build
build: ${BUILD_DIRS} ${BUILD}

${ENTRYPOINT}: ${ROOT_DIR}/resources/entrypoint
	cp $< $@

${BUILD}: ${GO_LIBS} ${SRC} ${ENTRYPOINT}
	docker buildx build      \
		--tag ${SERVICE}_build \
		--target service_build \
		${GLOBAL_BUILD_ARGS}   \
		--file ${DOCKERFILE}   \
		${DIR}
	touch $@

${TEST}: ${GO_LIBS} ${SRC}
	docker buildx build     \
		--tag ${SERVICE}_test \
		--target service_test \
		${GLOBAL_BUILD_ARGS}  \
		--file ${DOCKERFILE}  \
		${DIR}
	touch $@

.PHONY: create_service
create-service: test build ${CREATE_SERVICE}

${CREATE_SERVICE}: ${TEST} ${BUILD}
	docker buildx build    \
		--tag ${SERVICE}     \
		--target service     \
		--file ${DOCKERFILE} \
		${GLOBAL_BUILD_ARGS} \
		${DIR}
	touch $@

${GO_LIBS}:
	cp -a ${ROOT_DIR}/src/ ${GO_VENDOR}/landtitle

.PHONY: test
test: ${BUILD_DIRS} ${TEST}

${BUILD_DIRS}:
	mkdir -p $@
	#can't put GO_VENDER in BUILD_DIRS
	mkdir -p ${GO_VENDOR}

.PHONY: clean
clean:
	@rm -rf ${BUILD_DIRS} ${GO_VENDOR}

DIR := $(realpath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
ROOT_DIR ?= $(realpath ${DIR}/../..)
SCRIPTS_DIR ?= ${ROOT_DIR}/scripts
SRC_DIR := ${DIR}/src

BUILD_DIR := ${DIR}/build
BUILD_DIRS := ${BUILD_DIR}

BUILD := ${BUILD_DIR}/build
TEST := ${BUILD_DIR}/test

DOCKERFILE ?= ${DIR}/dockerfile

SERVICE ?= $(shell basename ${DIR})
SRC := $(shell find ${SRC_DIR} -type f -name '*.go') ${DOCKERFILE}
GO_VERSION ?= $(shell ${SCRIPTS_DIR}/scrape_go_version.sh ${DIR}/go.mod)

.PHONY: build
build: ${BUILD_DIRS} ${BUILD}

${BUILD}: ${SRC} ${BUILD_DIRS}
	docker build                             \
		--tag ${SERVICE}_build               \
		--target ${SERVICE}_build            \
		--build-arg GO_VERSION=${GO_VERSION} \
		${DIR}
	@touch $@

${TEST}: ${SRC} ${BUILD_DIRS}
	docker build                             \
		--tag ${SERVICE}_test                \
		--target ${SERVICE}_test             \
		--build-arg GO_VERSION=${GO_VERSION} \
		${DIR}

.PHONY: test
test: ${BUILD_DIRS} ${TEST}

${BUILD_DIRS}:
	@mkdir -p $@

.PHONY: clean
clean:
	@rm -rf ${BUILD_DIRS}

#TODO write a help target, generalize it as a tool for ALL makefiles?
#TODO write individual test with run support
DIR := $(realpath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
export ROOT_DIR := ${DIR}

SRC_DIR := ${DIR}/src
GO_VENDOR := ${SRC_DIR}/vendor
GO_SRC := $(shell find ${SRC_DIR} -type f -name '*.go')

BUILD_DIR := ${DIR}/build
BUILD_DIRS := ${BUILD_DIR}

export RESOURCE_DIR := ${DIR}/resources
export DOCKERFILE := ${RESOURCE_DIR}/Dockerfile

export SERVICE_DIR := ${DIR}/services
SERVICES := $(shell find ${SERVICE_DIR} -mindepth 1 -maxdepth 1 -type d -exec basename {} \;)

BUILD_SERVICE_TGTS := $(SERVICES:%=build-%-service)
TEST_SERVICE_TGTS := $(SERVICES:%=test-%-service)
CLEAN_SERVICE_TGTS := $(SERVICES:%=clean-%-service)
CREATE_SERVICE_TGTS := $(SERVICES:%=create-%-service)

LIB_MOD_DIRS := $(shell find ${SRC_DIR} -mindepth 1 -maxdepth 1 -type d)
LIB_MODS := $(foreach lib,${LIB_MOD_DIRS},$(shell basename $(lib)))
LIB_MOD_TESTS := $(LIB_MODS:%=${BUILD_DIR}/lib_%_mod_test)
LIB_SRC := $(shell find ${SRC_DIR} -type f -name '*.go')
CLEAN_LIB_TGTS := $(LIB_MODS:%=clean-%-lib)

export SCRIPTS_DIR := ${DIR}/scripts

LIB_GO_VERSION := $(shell ${SCRIPTS_DIR}/scrape_go_version.sh ${SRC_DIR}/go.mod)

GO_SRC := $(filter-out ${SRC_DIR}/vendor/%, $(shell find ${SRC_DIR} -type f -name '*.go'))

#TODO, should i somehow differntiate between go/no-go for tests and/or keep
#the test output for future reference between runs?
#TODO rename these to consecutive number extensions instead, see 
#scripts/rename_previous_files.sh for the ALMOST complete tool, wasting time
#do it later
define GET_LIB_SRC
$(1)_LIB_SRC := $$(shell find ${SRC_DIR}/$(1) -type f -name '*.go') $$(wildcard ${SRC_DIR}/$(1)/testdata/**/)
${BUILD_DIR}/lib_$(1)_mod_test: $${$(1)_LIB_SRC}
	@if test -e $$@; then mv $$@ $$@_$$(shell date '+%Y%m%d%H%M%S'); fi
	docker run --rm                                   \
		-v ${SRC_DIR}:/usr/src  \
		-v ${BUILD_DIR}:/output \
		-w /usr/src \
		golang:${LIB_GO_VERSION} \
		bash -c 'go test -v ./$(1)/... > /output/$$(notdir $$@)' || true
	@cat $$@
endef

TEST_LIBS := ${BUILD_DIR}/test_libs

$(foreach lib,${LIB_MODS},$(eval $(call GET_LIB_SRC,$(lib))))

ENV ?= local
include ${DIR}/${ENV}.mk

.PHONY: build
build: ${BUILD_SERVICE_TGTS}

.PHONY: create_services
create_services: ${CREATE_SERVICE_TGTS}

create-%-service: test_libs
	$(MAKE) -C ${SERVICE_DIR} $@

.PHONY: test
test: test_libs ${TEST_SERVICE_TGTS}

.PHONY: test_libs
test_libs: vendor_libs ${BUILD_DIRS} ${TEST_LIBS}

${TEST_LIBS}: ${GO_SRC}
	docker run --rm              \
		-v ${SRC_DIR}:/usr/src   \
		-w /usr/src              \
		golang:${LIB_GO_VERSION} \
		go test -v ./...
	@touch $@

#For development, only tests this particular directory
#@not sure why make is complaining about this empty rule, but it is
test-%-lib: ${BUILD_DIRS} vendor_libs ${BUILD_DIR}/lib_%_mod_test
	@echo -n "stfu" > /dev/null

.PHONY: clean
clean: ${CLEAN_SERVICE_TGTS} clean_vendor_libs
	@rm -rf ${BUILD_DIRS}

.PHONY: clean_services
clean_services: ${CLEAN_SERVICE_TGTS}

build-%-service: vendor_libs
	$(MAKE) -C ${SERVICE_DIR} $@

build_services: ${BUILD_SERVICE_TGTS}

test-%-service: vendor_libs
	$(MAKE) -C ${SERVICE_DIR} $@

.PHONY: clean_vendor_libs
clean_vendor_libs:
	@rm -rf ${GO_VENDOR}

.PHONY: vendor_libs
vendor_libs: ${GO_VENDOR}

${GO_VENDOR}:
	docker run --rm                \
		-v ${SRC_DIR}:/usr/src   \
		-w /usr/src              \
		golang:${LIB_GO_VERSION} \
		go mod vendor -v

.PHONY: test_services
test_services: ${TEST_SERVICE_TGTS}

clean-%-service:
	$(MAKE) -C ${SERVICE_DIR} $@

clean-%-lib:
	@rm -rf ${BUILD_DIR}/lib_$*_mod_test

${BUILD_DIRS}:
	@mkdir -p $@

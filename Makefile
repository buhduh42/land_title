#TODO create_services target doesn't seem to be respecting time stamps
#shouldn't be running everytime
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

define GET_LIB_SRC
$(1)_LIB_SRC := $$(shell find ${SRC_DIR}/$(1) -type f -name '*.go')
${BUILD_DIR}/lib_$(1)_mod_test: $${$(1)_LIB_SRC}
	docker build                                 \
		--tag $(1)_lib_test                      \
		--target lib_test                        \
		--build-arg GO_VERSION=${LIB_GO_VERSION} \
		--build-arg LIB=$(1)                     \
		--file ${DOCKERFILE}                     \
		${SRC_DIR}
	@touch $$@
endef

TEST_LIBS := ${BUILD_DIR}/test_libs

$(foreach lib,${LIB_MODS},$(eval $(call GET_LIB_SRC,$(lib))))

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

#TODO would this speed it up? probably not
${TEST_LIBS}: ${GO_SRC}
	docker run --rm              \
		-v ${SRC_DIR}:/usr/src   \
		-w /usr/src              \
		golang:${LIB_GO_VERSION} \
		go test -v ./...
	@touch $@

#For development, only tests this particular directory
test-%-lib: ${BUILD_DIRS} vendor_libs ${BUILD_DIR}/lib_%_mod_test
	@echo -n ""

.PHONY: clean
clean: ${CLEAN_SERVICE_TGTS}
	@rm -rf ${BUILD_DIRS} ${GO_VENDOR}

.PHONY: clean_services
clean_services: ${CLEAN_SERVICE_TGTS}

build-%-service: vendor_libs
	$(MAKE) -C ${SERVICE_DIR} $@

build_services: ${BUILD_SERVICE_TGTS}

test-%-service: vendor_libs
	$(MAKE) -C ${SERVICE_DIR} $@

.PHONY: vendor_libs
vendor_libs: ${GO_VENDOR}

${GO_VENDOR}:
	docker run --rm              \
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

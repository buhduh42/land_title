DIR := $(realpath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
export ROOT_DIR := ${DIR}
SERVICE_DIR := ${DIR}/services
SERVICES := $(wildcard ${SERVICE_DIR}/*)

BUILD_SERVICE_TGS := $(SERVICES:%=build-%-service)
TEST_SERVICE_TGS := $(SERVICES:%=test-%-service)
CLEAN_SERVICE_TGS := $(SERVICES:%=clean-%-service)

export SCRIPTS_DIR := ${DIR}/scripts

.PHONY: build
build: ${BUILD_SERVICE_TGTS}

.PHONY: clean
clean: ${CLEAN_SERVICE_TGTS}

build-%-service:
	$(MAKE) -C ${SERVICE_DIR} $@

test-%-service:
	$(MAKE) -C ${SERVICE_DIR} $@

clean-%-service:
	$(MAKE) -C ${SERVICE_DIR} $@

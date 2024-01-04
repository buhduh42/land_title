#TODO Will likely include or just get rid of the subdirectory makefiles here
DIR := $(realpath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

build-%-service:
	$(MAKE) -C ${DIR}/$* build

test-%-service:
	$(MAKE) -C ${DIR}/$* test

clean-%-service:
	$(MAKE) -C ${DIR}/$* clean

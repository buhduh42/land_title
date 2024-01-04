### Makefile Reference
#### Targets:
test-%-lib -- Each child directory in the src/ directory is a "lib".
This will invoke the "lib_test" target in the Dockerfile for the matched lib.

test_libs -- Runs all tests for libs, see the above target.

test-%-service -- Each child directory in the services/ directory is a service.  
This will invoke the service_test target in the Dockerfile for the matched service.

test_services -- Runs all tests for services, see the above target.

build-%-service -- Builds a docker image whose name is %_service, intended to be used in the cluster.

build_services -- Runs all of the above images.

clean-%-lib -- Removes build artifacts for the matched lib.  Probably just an empty file in the build/ directory

clean_libs -- Cleans all lib artifacts, see the above target.

clean-%-service -- Cleans the matched service artifacts.

clean_services -- Cleans all service artifacts, see the above target.

clean -- Cleans ALL build artifacts and deletes all "build" directories.
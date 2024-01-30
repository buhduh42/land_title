#!/bin/bash
if ! test ${TEST_INSTRUCTIONS}; then
  exit 0
fi

if ! test $1; then
  >&2 echo "test_name or test_indeces required as argument to this script"
  exit 1
fi

if test "$1" = "test_name"; then
  output=$(echo -n ${TEST_INSTRUCTIONS} | cut -f1 -d:)
elif test "$1" = "test_indeces"; then
  output=$(echo -n ${TEST_INSTRUCTIONS} | cut -f2- -d:)
else
  >&2 echo "test_name or test_indeces required as argument to this script"
  exit 1
fi

if test -z ${output}; then
  exit 0
fi

echo -n ${output}

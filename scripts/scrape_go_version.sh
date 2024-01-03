#!/bin/bash

if test -z $1 || ! test -f $1; then
  >&2 echo "pass a go.mod file as an argument to this script"
  exit 1
fi

version=$(\
  grep -m 1 -x 'go[[:space:]]\+1\.[[:digit:]]\+\(\.[[:digit:]]\+\)\?' $1 \
  | sed 's/^go[[:space:]]\(1\.[[:digit:]]\+\(\.[[:digit:]]\+\)\?\)$/\1/')

if test -z ${version}; then
  2>&1 echo "unable to parse go version from $1"
  exit 1
fi

echo -n ${version}

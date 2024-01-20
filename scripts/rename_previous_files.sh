#!/bin/bash

if ! test -e $1; then
  exit 0
fi

root=$1
len=$(( ${#root}-4 ))

if test "${root:${len}:4}" = ".bak"; then
  exit 0
fi

toChange=$(find $(dirname ${root}) -name "$(basename ${root})*" -type f -exec echo {} \;)

sorted=$(echo ${toChange} | sort -r -)

for f in ${sorted[@]}; do
  check=${f%.*}
  if echo ${f} -n | grep -q "${check}\.[[:digit:]]\+" -; then
    num=$(echo -n ${f} | sed 's/.\+\.\([[:digit:]]\+\)$/\1/')
    mv ${f} ${check}.$(( ${num}+1 ))
  fi
done

mv ${root} ${root}.1

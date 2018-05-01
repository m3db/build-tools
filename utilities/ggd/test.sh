#!/bin/bash
#
# Copyright (c) 2018 Uber Technologies, Inc.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

set -e # paranoia, ftw

function banner() {
  local msg=$1
  echo ; echo ;
  echo "###############################"
  echo "${msg}"
  echo "###############################"
}

# NB: need to set these or `git init` fails in the CI.
[ ! -z "$(git config --global user.name)" ] || \
  git config --global user.name "CIRunner"
[ ! -z "$(git config --global user.email)" ] || \
  git config --global user.email "ci@ci.com"

cwd=${GOPATH}/src/github.com/m3db/build-tools
temp_dir=$(mktemp -d)

echo "Creating stub directories at ${temp_dir}"
mkdir -p ${temp_dir}/bin
mkdir -p ${temp_dir}/src/github.com/testcase

echo "Building ggd and moving"
mkdir -p ${cwd}/bin
go build .
mv ggd ${temp_dir}/bin/.

# think of this as a defer func() in golang
function defer {
  echo "Cleaning up directories under ${temp_dir}"
  rm -rf ${temp_dir}
}
trap defer EXIT

for testfile in $(ls ${cwd}/utilities/ggd/tests | grep 9); do
  filename=$(basename -- "$testfile")
  test_case="${filename%.*}"
  banner "${test_case} starting test case"
  echo "${test_case} creating gh directory"
  rm -rf ${temp_dir}/src/github.com/testcase/testcase
  mkdir -p ${temp_dir}/src/github.com/testcase/testcase

  echo "${test_case} wiring up master"
  cp -r ${cwd}/utilities/ggd/testdata/${test_case}/master/* ${temp_dir}/src/github.com/testcase/testcase/
  cd ${temp_dir}/src/github.com/testcase/testcase
  git init .
  git add .
  git commit -m 'pushing to master'

  echo "${test_case} wiring up branch"
  git checkout -b branch
  rm -rf ${temp_dir}/src/github.com/testcase/testcase/*
  cp -r ${cwd}/utilities/ggd/testdata/${test_case}/branch/* ${temp_dir}/src/github.com/testcase/testcase/
  git add .
  git commit -m 'pushing to branch'

  mkdir -p ${cwd}/utilities/ggd/bin
  TARGET_FILE=${cwd}/utilities/ggd/bin/${test_case}.out
  echo "${test_case} executing test"
  GOPATH=${temp_dir} PATH=${temp_dir}/bin:$PATH \
    ${cwd}/utilities/ggd/tests/${testfile} >${TARGET_FILE} 2>&1

  diff ${cwd}/utilities/ggd/testdata/${test_case}/expected.out ${TARGET_FILE} || \
    (echo "${test_case} failed, files differ"; exit 1)

  echo "${test_case} finished successfully!"
  echo "${test_case} cleaning up"
done

banner "all tests finished successfully!!!"
echo ; echo
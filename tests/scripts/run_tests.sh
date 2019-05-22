#!/bin/bash
set -ex

export GOPATH=/tmp/go
export VENDORDIR=github.com/MottainaiCI
SUBPROJECT="${SUBPROJECT:-mottainai-server}"
PROJECT_DIR="${PROJECT_DIR:-tests}"

git clone https://github.com/kward/shunit2.git /tmp/unit 2>&1 >/dev/null
cp -rf ${GOPATH}/src/${VENDORDIR}/${SUBPROJECT}/$PROJECT_DIR/suite /tmp/unit
pushd /tmp/unit/suite/ 2>&1 >/dev/null

echo
echo
echo
echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
echo "Running tests"
echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
echo "@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@"
bash -ex run.sh
popd 2>&1 >/dev/null

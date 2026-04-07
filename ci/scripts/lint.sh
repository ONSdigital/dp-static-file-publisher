#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-static-file-publisher
  make lint
popd

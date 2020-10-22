#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-static-file-publisher
  make audit
popd
#!/bin/bash -eux

pushd dp-static-file-publisher
  make build
  cp build/dp-static-file-publisher Dockerfile.concourse ../build
popd

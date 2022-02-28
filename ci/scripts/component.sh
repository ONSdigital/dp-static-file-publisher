#!/bin/bash -eux

pushd dp-static-file-publisher
  make docker-test-component
popd

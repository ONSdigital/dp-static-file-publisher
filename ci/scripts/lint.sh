#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-static-file-publisher
# Install golangci-lint
  go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.45.2
  make lint
popd

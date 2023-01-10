#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-static-file-publisher
# Install golangci-lint
  go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1
  make lint
popd

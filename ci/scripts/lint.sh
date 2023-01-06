#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/log.go
# Install golangci-lint
  go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.41.1
  make lint
popd

#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-static-file-publisher
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2
  make lint
popd

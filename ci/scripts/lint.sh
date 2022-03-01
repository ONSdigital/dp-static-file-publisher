#!/bin/bash -eux

pushd dp-static-file-publisher
  go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0
  make lint
popd
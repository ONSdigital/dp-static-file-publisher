#!/bin/bash
set -x
awslocal s3 mb s3://public
awslocal s3 mb s3://private
set +x
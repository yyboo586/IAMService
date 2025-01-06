#!/bin/bash
set -e

# 安装golangci-lint(如果没有安装)
if ! command -v golangci-lint &> /dev/null; then
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
fi

# 运行lint检查
golangci-lint run ./... 
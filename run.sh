#!/bin/bash

set -e # 设置错误处理

echo "开始构建IAMService"
go mod tidy

echo "1、执行单元测试"
go generate ./... && go test ./...

echo "2、执行代码检查"
golangci-lint run ./...

echo "3、构建可执行文件"
go build -o ./bin/IAMService main.go

echo "4、IAMService构建完成，开始构建Docker镜像"
docker build -t iamservice:latest .

echo "5、删除已存在的IAMService容器"
docker rm -f iamservice || true

echo "6、启动依赖服务"
./scripts/deploy_nsq.sh

echo "7、启动IAMService服务镜像"
docker run -d --name iamservice \
    --network=host \
    iamservice:latest

echo "8、查看服务运行日志"
docker logs -f $(docker ps -a | grep iamservice | awk '{print $1}')
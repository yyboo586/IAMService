#!/bin/bash

go build -o ./bin/IAMService main.go

docker build -t iamservice:latest .

docker rm -f iamservice

# 使用host网络模式时，容器可以直接使用宿主机的端口
docker run -d --name iamservice --network=host iamservice:latest

docker logs -f $(docker ps -a | grep iamservice | awk '{print $1}')
#!/bin/bash

# 停止已存在的NSQ相关容器
docker rm -f nsqlookupd nsqd nsqadmin || true

# 启动nsqlookupd
docker run -d --name nsqlookupd \
    --network host \
    nsqio/nsq /nsqlookupd

# 等待nsqlookupd启动
sleep 2

# 启动nsqd
docker run -d --name nsqd \
    --network host \
    nsqio/nsq /nsqd \
    --broadcast-address=localhost \
    --lookupd-tcp-address=localhost:4160

# 启动nsqadmin
docker run -d --name nsqadmin \
    --network host \
    nsqio/nsq /nsqadmin \
    --lookupd-http-address=localhost:4161

echo "NSQ 部署完成！"
echo "nsqlookupd TCP: localhost:4160"
echo "nsqlookupd HTTP: localhost:4161"
echo "nsqd TCP: localhost:4150"
echo "nsqd HTTP: localhost:4151"
echo "nsqadmin HTTP: localhost:4171" 

# 命令行订阅topic
# docker run --network host nsqio/nsq /nsq_tail --topic=user_created --lookupd-http-address=localhost:4161
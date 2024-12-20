FROM ubuntu:22.04

WORKDIR /usr/local/bin

COPY ./bin/IAMService .

CMD ["./IAMService"]

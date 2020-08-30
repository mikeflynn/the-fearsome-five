FROM golang:1.14-alpine

RUN mkdir -p /go/src/the-fearsome-five && \
    apk update && \
    apk add git openssl

WORKDIR /go/src/the-fearsome-five

COPY . .

RUN openssl req -x509 -nodes -newkey rsa:4096 -subj "/C=CA/ST=California/L=Los Angeles/O=McDuck Inc/OU=Security/CN=localhost" -keyout server.key -out server.crt && \
    go build -o server-docker

EXPOSE 8888

CMD ./server-docker -verbose
FROM golang:1.14-alpine

RUN mkdir -p /go/src/the-fearsome-five && \
    apk update && \
    apk add git

WORKDIR /go/src/the-fearsome-five

COPY . .

RUN go build -o client-docker

EXPOSE 8888

CMD ./client-docker -verbose -delay=10 -server=localhost:8080 -reset
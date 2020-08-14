FROM golang:1.14-alpine

RUN mkdir -p /go/src/the-fearsome-five && \
    apk update && \
    apk add gcc

WORKDIR /go/src/the-fearsome-five

COPY . .

RUN cd client && go build -o client-docker && \
    cd server && go build -o server-docker

EXPOSE 8888

CMD ./docker-start.sh
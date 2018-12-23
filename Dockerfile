# frontend server image
FROM nginx:alpine as frontend 

WORKDIR /web

COPY ./frontend/assets .
# Overwrite the default nginx config file
COPY ./config/nginx.conf /etc/nginx/


# data store server image
FROM redis:alpine as redis

WORKDIR /redis
COPY ./config/redis.conf .


# application builder image
FROM golang:alpine as builder

WORKDIR /go/src/github.com/wuyrush/snippet

# install git to use go get to add application dependencies
RUN apk update; \
    apk add git && echo "Git installed";    \
    go get -d -v github.com/satori/go.uuid  \
        github.com/Sirupsen/logrus  \
        github.com/go-redis/redis \
        github.com/kelseyhightower/envconfig \
        github.com/gorilla/mux

COPY ./*.go ./ 

RUN GOOS=linux go build -v -o ./build/app .

# Application server image
# Use a new image to include our final executable ONLY, resulting tiny image size
FROM alpine:latest as backend

WORKDIR /snippet

COPY --from=builder /go/src/github.com/wuyrush/snippet/build/app .
# note we can pass args via `docker run` if we use the exec form of ENTRYPOINT
ENTRYPOINT ["./app"]

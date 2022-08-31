# syntax=docker/dockerfile:1
FROM golang:1.19.0 AS build
WORKDIR /src
ADD . /src
RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X main.buildSha=`git rev-parse HEAD` -X main.buildTime=`date +'%Y-%m-%d_%T'`"

FROM alpine:3.16.0
RUN addgroup -S -g 1001 appgroup && adduser -S -u 1001 -G appgroup appuser
USER appuser
COPY --from=build /src/k8s-ecr-login-renew /k8s-ecr-login-renew
CMD ["/k8s-ecr-login-renew"]

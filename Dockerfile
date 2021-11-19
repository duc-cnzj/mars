FROM node:lts-alpine as web-build

WORKDIR /app

COPY . .

RUN cd frontend && \
    yarn install --registry=https://registry.npm.taobao.org && \
    yarn build

FROM golang:1.17-alpine3.14 AS builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
  apk add --no-cache ca-certificates tzdata build-base

#ARG UPX_VERSION=3.96
#
#ADD https://github.com/upx/upx/releases/download/v${UPX_VERSION}/upx-${UPX_VERSION}-amd64_linux.tar.xz /tmp/upx.tar.xy
#RUN tar -xJOf /tmp/upx.tar.xy upx-${UPX_VERSION}-amd64_linux/upx > /usr/local/bin/upx \ 
# && chmod +x /usr/local/bin/upx

WORKDIR /app

COPY . .

COPY --from=web-build /app/frontend/build /app/frontend/build

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod download

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /bin/app main.go
#  && upx -9 /bin/app

FROM alpine:3.14

WORKDIR /

COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/app /bin/app

CMD ["app"]

FROM node:lts-alpine as web-build

WORKDIR /app

COPY . .

RUN cd frontend && \
    yarn install --registry=https://registry.npm.taobao.org && \
    yarn build

FROM golang:1.16-alpine3.13 AS builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
  apk add --no-cache ca-certificates tzdata build-base

WORKDIR /app

COPY . .

COPY --from=web-build /app/frontend/build /app/frontend/build

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod download

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /bin/app main.go

FROM alpine

WORKDIR /

COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/app /bin/app

CMD ["app"]

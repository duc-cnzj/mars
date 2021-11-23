FROM node:lts-alpine as web-build

WORKDIR /app

COPY . .

RUN cd frontend && \
    yarn install --registry=https://registry.npm.taobao.org && \
    yarn build

FROM golang:1.17-alpine3.14 AS builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
  apk add --no-cache ca-certificates tzdata build-base git

WORKDIR /app

COPY . .

COPY --from=web-build /app/frontend/build /app/frontend/build

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod download

RUN VERSION_PATH=$(go list -m -f "{{.Path}}")/version && LDFLAGS="-w -s  \
     -X ${VERSION_PATH}.gitRepo=$(go list -m -f "{{.Path}}")
     -X ${VERSION_PATH}.gitBranch=$(git rev-parse --abbrev-ref HEAD) \
     -X ${VERSION_PATH}.buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
     -X ${VERSION_PATH}.gitCommit=$(git rev-parse --short HEAD) \
     -X ${VERSION_PATH}.gitTag=$(git describe --exact-match --tags HEAD 2> /dev/null || echo "") \
     -X ${VERSION_PATH}.kubectlVersion=$(go list -m -f "{{.Path}} {{.Version}}" all | grep k8s.io/client-go | cut -d " " -f2) \
     -X ${VERSION_PATH}.helmVersion=$(go list -m -f "{{.Path}} {{.Version}}" all | grep helm.sh/helm/v3 | cut -d " " -f2)" \
    && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o /bin/app main.go

FROM alpine:3.14

WORKDIR /

COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/app /bin/app

CMD ["app"]

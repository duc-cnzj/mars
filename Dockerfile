FROM --platform=$TARGETPLATFORM node:lts-alpine as web-build

WORKDIR /app

COPY . .

RUN cd frontend && \
    yarn install --registry=https://registry.npm.taobao.org && \
    yarn build

FROM --platform=$TARGETPLATFORM golang:1.17 AS builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apt/sources.list && \
  apt install -y ca-certificates tzdata git gcc-aarch64-linux-gnu

WORKDIR /app

COPY . .

COPY --from=web-build /app/frontend/build /app/frontend/build

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod download

RUN if [ "$TARGETARCH" = "arm64" ]; then CC=aarch64-linux-gnu-gcc && CC_FOR_TARGET=gcc-aarch64-linux-gnu && EXTRA_FLAGS='-linkmode external -extldflags "-static"'; fi && \
    VERSION_PATH=$(go list -m -f "{{.Path}}")/version && LDFLAGS="-w -s  \
     -X ${VERSION_PATH}.gitRepo=$(go list -m -f '{{.Path}}') \
     -X ${VERSION_PATH}.gitBranch=$(git rev-parse --abbrev-ref HEAD) \
     -X ${VERSION_PATH}.buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
     -X ${VERSION_PATH}.gitCommit=$(git rev-parse --short HEAD) \
     -X ${VERSION_PATH}.gitTag=$(git describe --exact-match --tags HEAD 2> /dev/null || echo '') \
     -X ${VERSION_PATH}.kubectlVersion=$(go list -m -f '{{.Path}} {{.Version}}' all | grep k8s.io/client-go | cut -d ' ' -f2) \
     -X ${VERSION_PATH}.helmVersion=$(go list -m -f '{{.Path}} {{.Version}}' all | grep helm.sh/helm/v3 | cut -d ' ' -f2)" \
    && CGO_ENABLED=1 CC=$CC CC_FOR_TARGET=$CC_FOR_TARGET GOOS=$TARGETPLATFORM GOARCH=$TARGETARCH go build -ldflags="$LDFLAGS $EXTRA_FLAGS" -o /bin/app main.go

FROM --platform=$TARGETPLATFORM alpine:3.14

WORKDIR /

COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/app /bin/app

CMD ["app"]

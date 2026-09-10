FROM --platform=$BUILDPLATFORM node:22-alpine AS web-build

WORKDIR /app

COPY ./frontend .

RUN yarn install --registry=https://registry.npm.taobao.org && \
    yarn build

FROM --platform=linux/amd64 golang:1.25 AS builder

ARG TARGETARCH
ARG TARGETOS

RUN apt update && \
  apt install -y ca-certificates tzdata git gcc-aarch64-linux-gnu xz-utils && \
  wget $(curl -s https://api.github.com/repos/upx/upx/releases/latest \
    | grep browser_download_url | grep amd64 | cut -d '"' -f 4) -O upx.tar.xz && \
  tar -xvf upx.tar.xz && \
  cd upx-*-amd64_linux && \
  mv upx /bin/upx

WORKDIR /app

# 先只拷模块描述文件，让 go mod download 单独成层：只依赖 go.mod/go.sum/go.work*，
# 源码变更（含每次提交都会变的 .git）不会让它失效，本地重复构建可复用该层。
# 已验证：仅凭这 6 个文件即可枚举出全部 942 个模块，与含完整源码时结果一致。
COPY go.mod go.sum go.work go.work.sum ./
COPY api/go.mod api/go.sum ./api/

RUN go mod download

COPY . .

COPY --from=web-build /app/build /app/frontend/build

# tag 触发的 CI 构建里 actions/checkout 检出的是 detached HEAD，
# `git rev-parse --abbrev-ref HEAD` 会返回字面量 "HEAD"，而该值经 services/version
# 直接暴露给用户，必须修正。CI 用 build-arg GIT_BRANCH 传入真实 ref 名；
# 本地构建没有该 arg 时回退到 git 命令，回退结果仍是 "HEAD" 则用精确 tag 名兜底，
# 保证任何路径都不会把 "HEAD" 写进版本信息。
# 注意 ARG 必须放在 COPY . . 之后：放在前面会让 build-arg 变化成为缓存键的一部分，
# 反而白白打断前面的层缓存。
ARG GIT_BRANCH

# 编译缓存挡在层外：/root/.cache/go-build 是 GB 级的工具缓存，写进镜像层只会白白
# 撑大镜像，用 cache mount 承载即可让本地重复构建复用，不进最终产物。
# 注意这里刻意不给 /go/pkg/mod 挂 cache mount：模块缓存要靠上层 go mod download 的
# 镜像层携带而来，挂了 cache mount 反而会遮住层里的内容，退化成每次重新下载。
RUN --mount=type=cache,target=/root/.cache/go-build \
    if [ "$TARGETARCH" = "arm64" ]; then CC=aarch64-linux-gnu-gcc && CC_FOR_TARGET=gcc-aarch64-linux-gnu && EXTRA_FLAGS='-extldflags "-static"'; fi && \
    VERSION_PATH=$(go list -m -f "{{.Path}}" | grep -v api)/internal/version && \
    BUILD_BRANCH="${GIT_BRANCH}"; \
    if [ -z "$BUILD_BRANCH" ]; then BUILD_BRANCH=$(git rev-parse --abbrev-ref HEAD); fi; \
    if [ "$BUILD_BRANCH" = "HEAD" ]; then BUILD_BRANCH=$(git describe --exact-match --tags HEAD 2> /dev/null || echo '<unknown>'); fi; \
    LDFLAGS="-w -s  \
     -X ${VERSION_PATH}.gitBranch=${BUILD_BRANCH} \
     -X ${VERSION_PATH}.buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
     -X ${VERSION_PATH}.gitCommit=$(git rev-parse --short HEAD) \
     -X ${VERSION_PATH}.gitTag=$(git describe --exact-match --tags HEAD 2> /dev/null || echo '') \
     -X ${VERSION_PATH}.kubectlVersion=$(go list -m -f '{{.Path}} {{.Version}}' all | grep k8s.io/client-go | cut -d ' ' -f2) \
     -X ${VERSION_PATH}.helmVersion=$(go list -m -f '{{.Path}} {{.Version}}' all | grep helm.sh/helm/v3 | cut -d ' ' -f2)" \
    && CGO_ENABLED=1 CC=$CC CC_FOR_TARGET=$CC_FOR_TARGET GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="$LDFLAGS $EXTRA_FLAGS" -o /bin/app main.go \
    && upx -9 /bin/app

FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=builder /bin/app /bin/app

CMD ["app"]

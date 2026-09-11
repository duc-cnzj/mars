VERSION_PATH=$(shell go list -m -f "{{.Path}}" | grep -v api)/internal/version
PROTO_FILES=$(shell find api -name *.proto)
VERSION=$(shell git describe --exact-match --tags HEAD 2> /dev/null || echo "")
LDFLAGS=-w -s  \
 -X ${VERSION_PATH}.gitBranch=$(shell git rev-parse --abbrev-ref HEAD) \
 -X ${VERSION_PATH}.buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
 -X ${VERSION_PATH}.gitCommit=$(shell git rev-parse --short HEAD) \
 -X ${VERSION_PATH}.gitTag=${VERSION} \
 -X ${VERSION_PATH}.kubectlVersion=$(shell go list -m -f "{{.Path}} {{.Version}}" all | grep k8s.io/client-go | cut -d " " -f2) \
 -X ${VERSION_PATH}.helmVersion=$(shell go list -m -f "{{.Path}} {{.Version}}" all | grep helm.sh/helm/v3 | cut -d " " -f2)

# 构建产物统一输出到 bin/（.gitignore 已忽略，不进入版本库）
BIN_DIR ?= bin

$(BIN_DIR):
	mkdir -p $@

# protoc 本体是 C++ 二进制（go install 装不了），这里钉版本装到项目内 bin/protoc/，
# 与系统隔离、免 sudo，且 include/（well-known types）随包自带、protoc 按自身相对路径自动找到。
# 改版本号即可；可选版本：https://github.com/protocolbuffers/protobuf/releases
PROTOC_VERSION ?= 31.1
PROTOC := $(BIN_DIR)/protoc/bin/protoc

# 平台后缀映射（uname → GitHub release 命名；Windows 的 protoc-gen 建议走 WSL/Docker）
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
PROTOC_OS := $(if $(findstring Darwin,$(UNAME_S)),osx,$(if $(findstring Linux,$(UNAME_S)),linux,win))
PROTOC_ARCH := $(if $(findstring arm64,$(UNAME_M)),aarch_64,x86_64)
PROTOC_SUFFIX := $(if $(filter win,$(PROTOC_OS)),win64,$(PROTOC_OS)-$(PROTOC_ARCH))

$(PROTOC): | $(BIN_DIR)
	mkdir -p $(BIN_DIR)/protoc
	curl -fLo $(BIN_DIR)/protoc.zip \
		https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_SUFFIX).zip
	unzip -oq $(BIN_DIR)/protoc.zip -d $(BIN_DIR)/protoc
	rm $(BIN_DIR)/protoc.zip

.PHONY: protoc
protoc: $(PROTOC)

# go install 的落地目录（GOBIN 未设置时回退 GOPATH/bin）；自检按同一路径读二进制，
# 否则"钉了版本"和"protoc 实际启动的插件"可能是两个东西。
TOOLS_BIN := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)

# 生成器钉版本（安装与自检共用同一批变量，杜绝两处版本号各写一遍再次漂移）
PGV_VERSION     := v1.3.3
GATEWAY_VERSION := v2.30.0
GO_GRPC_VERSION := v1.6.2
GO_VERSION      := v1.36.12
OPENAPI_VERSION := v0.7.0

.PHONY: build_tools
# 生成器版本必须显式钉死：不带 @version 的 go install 按主模块 MVS 解析，会被传递依赖顶高，
# 导致生成器与 go.mod 声明漂移、产物不可复现。钉版本是产物可复现的唯一事实来源——
# 改这里必须重生成 api/ 下产物并一起提交，CI 不跑 make api，漂移没人替你发现。
# 四个 Go 插件版本与 go.mod 声明对齐（protobuf v1.36.12 / grpc-gateway v2.30.0 /
# protoc-gen-validate v1.3.3；protoc-gen-go-grpc 独立模块，版本与 grpc 运行时无关）。
# protoc-gen-openapi(gnostic) 被 kube-openapi 钳在 v0.7.0，升 v0.7.1 会换 yaml module path 断 build，勿动。
build_tools:
	# go install 一次调用要求所有参数同模块同版本，各工具版本不同，必须拆开
	go install github.com/envoyproxy/protoc-gen-validate@$(PGV_VERSION)
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(GATEWAY_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(GO_GRPC_VERSION)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(GO_VERSION)
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@$(OPENAPI_VERSION)
	@echo "== 工具链版本自检 =="
	# 不用插件各自的 --version：protoc-gen-validate / gnostic 的 protoc-gen-openapi 都不支持该参数，
	# 其余输出格式也不统一（有的带 v 前缀有的不带）。二进制内嵌的 module path+version 是唯一统一口径。
	# PATH 断言同样承重：protoc 按 PATH 找插件，钉版本装在别处等于没钉。
	@fail=0; for spec in \
		protoc-gen-validate:github.com/envoyproxy/protoc-gen-validate:$(PGV_VERSION) \
		protoc-gen-grpc-gateway:github.com/grpc-ecosystem/grpc-gateway/v2:$(GATEWAY_VERSION) \
		protoc-gen-go-grpc:google.golang.org/grpc/cmd/protoc-gen-go-grpc:$(GO_GRPC_VERSION) \
		protoc-gen-go:google.golang.org/protobuf:$(GO_VERSION) \
		protoc-gen-openapi:github.com/google/gnostic:$(OPENAPI_VERSION) ; do \
		bin="$${spec%%:*}"; rest="$${spec#*:}"; mod="$${rest%:*}"; want="$${rest##*:}"; \
		path="$(TOOLS_BIN)/$$bin"; \
		got="$$(go version -m "$$path" 2>/dev/null | awk -v m="$$mod" '$$1=="mod" && $$2==m {print $$3; exit}')"; \
		onpath="$$(command -v $$bin 2>/dev/null || true)"; \
		if [ "$$got" = "$$want" ] && [ "$$onpath" = "$$path" ]; then \
			echo "  [ok]   $$bin $$got"; \
		else \
			echo "  [FAIL] $$bin 钉版本 $$want / 实际 '$$got' / PATH 解析 '$$onpath'"; fail=1; \
		fi; \
	done; \
	if [ "$$fail" -ne 0 ]; then \
		echo "工具链自检失败：生效版本与钉版本不一致，产物将不可复现"; exit 1; \
	fi

.PHONY: api
api: $(PROTOC)
	$(PROTOC) \
        --proto_path=./api \
		--proto_path ./third_party/protos \
		--go_out=paths=source_relative:./api \
		--go-grpc_out=paths=source_relative:./api \
		--grpc-gateway_out=paths=source_relative:./api \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt paths=source_relative \
		--grpc-gateway_opt generate_unbound_methods=true \
		--validate_out=lang=go,paths=source_relative:./api \
	    --openapi_out=enum_type=string,fq_schema_naming=true,default_response=true,version="$(VERSION)",title="mars api.":./doc \
		$(PROTO_FILES)

	npx openapi-typescript ./doc/openapi.yaml --enum --enum-values --properties-required-by-default -o ./frontend/src/api/schema.d.ts

	./frontend/node_modules/.bin/pbjs -t static-module -o ./frontend/src/api/websocket.js -w es6  ./api/proto/websocket/websocket.proto  \
      --no-verify \
      --no-convert \
      --no-create \
      --force-number \
      --force-message \
      --no-delimited
#      --keep-case \
    #  --no-encode \
    #  --no-decode \

    # https://github.com/protobufjs/protobuf.js/blob/master/cli/README.md#reflection-vs-static-code
    #  Static targets only:
    #
    #  --no-create      Does not generate create functions used for reflection compatibility.
    #  --no-encode      Does not generate encode functions.
    #  --no-decode      Does not generate decode functions.
    #  --no-verify      Does not generate verify functions.
    #  --no-convert     Does not generate convert functions like from/toObject
    #  --no-delimited   Does not generate delimited encode/decode functions.
    #  --no-beautify    Does not beautify generated code.
    #  --no-comments    Does not output any JSDoc comments.
    #  --no-service     Does not output service classes.
    #
    #  --force-long     Enforces the use of 'Long' for s-/u-/int64 and s-/fixed64 fields.
    #  --force-number   Enforces the use of 'number' for s-/u-/int64 and s-/fixed64 fields.
    #  --force-message  Enforces the use of message instances instead of plain objects.

	./frontend/node_modules/.bin/pbts -o ./frontend/src/api/websocket.d.ts ./frontend/src/api/websocket.js --keep-case

	# 重新生成 HTTP SDK stub：proto 一改，httpclient 必须同步（防漂移）
	cd ./api && go run ./http/gen/cmd

.PHONY: clear_proto
clear_proto:
	rm -rf ./api/**/*.go

.PHONY: gen
gen:
	GOWORK=off go generate ./...

.PHONY: all
all: api gen fmt

.PHONY: wire
wire:
	cd ./cmd && go tool wire

.PHONY: sec
# api/ 是独立 Go 模块（go.work use api），根模块上下文扫描会因跨模块
# internal/flight 报 Golang errors 恒 exit 2，故排除 api/。
# G117 是 gosec 新规则：配置结构体里名为 password/client_secret 的字段
# 属正常运行时凭据载体，非硬编码秘密，排除。
# G120 是 gosec v2.26 新规则（无界 ParseMultipartForm）：唯一触发点是
# file_handler 上传接口，其上限来自 config.MaxUploadSize()（解析 UploadMaxSize，
# 失败回退 50M 下限），恒有界，排除。
sec:
	go tool gosec -exclude-dir=api -exclude=G104,G304,G115,G117,G120 -stdout -tests=false -exclude-generated -fmt=json -out=gosec-results.json  ./...

.PHONY: lint
lint:
	go tool golangci-lint run -D errcheck

.PHONY: release
release: build_linux_amd64 build_darwin_amd64 build_darwin_arm64

.PHONY: fmt
fmt:
	gofmt -s -w ./api && \
	gofmt -s -w -r 'interface{} -> any' ./internal ./third_party ./cmd && \
	go tool goimports -w ./

.PHONY: serve
serve:
	go run main.go serve
#	go run -race main.go serve --debug

# dev/ 基础设施一键启停（docker compose v2；dev/docker-compose.yml 只含依赖，不含 mars 本体）
COMPOSE_CMD := docker compose -f dev/docker-compose.yml

.PHONY: dev-up
# make dev-up：整栈后台启动（redis/mysql/minio/nsq/jaeger）
dev-up:
	$(COMPOSE_CMD) up -d

.PHONY: dev-down
# make dev-down：整栈关闭
dev-down:
	$(COMPOSE_CMD) down

.PHONY: dev-logs
# make dev-logs：跟随打印整栈日志（最近 100 行）
dev-logs:
	$(COMPOSE_CMD) logs -f --tail=100

.PHONY: dc-up
# make dc-up SVC=redis：只启动指定依赖服务；缺省 SVC 时等价 dev-up
dc-up:
	$(COMPOSE_CMD) up -d $(SVC)

.PHONY: dc-down
# make dc-down SVC=redis：只停止指定依赖服务；缺省 SVC 时等价 dev-down
dc-down:
	$(COMPOSE_CMD) down $(SVC)

.PHONY: build_race
build_race: | $(BIN_DIR)
	CGO_ENABLED=1 go build -ldflags="${LDFLAGS}" -race -o $(BIN_DIR)/app main.go

.PHONY: build
build: | $(BIN_DIR)
	CGO_ENABLED=1 go build -ldflags="${LDFLAGS}" -o $(BIN_DIR)/app main.go

.PHONY: build_web
build_web:
	cd ./frontend && yarn build

.PHONY: test
test:
	# 过滤自动生成代码（ent 生成代码 + mock 文件），仅保留手写业务代码参与覆盖率统计，
	# 否则生成的 ORM/桩代码（覆盖率恒 0%）会稀释真实覆盖率（96.7% → 24.6%）
	go test ./internal/... -race -count=1 -cover -coverprofile=coverage.txt -covermode atomic && \
	grep -v -e 'internal/data/ent/' -e 'mock_' coverage.txt > coverage.txt.filtered && \
	mv coverage.txt.filtered coverage.txt && \
	go tool cover -func coverage.txt

# test_all 是 CI 的完整门禁：先跑 test 拿到 internal 覆盖率（供 codecov 上传），
# 再补跑此前完全没有 CI 信号的包——根模块里除 internal/* 外的全部包
# （cmd/doc/examples/frontend/third_party；frontend 依赖 //go:embed build/*，CI 需先造
# stub，见 .github/workflows/test.yaml 的 Setup 步骤），以及独立模块 api/
# （go.work 的第二模块，根模块的 ./... 覆盖不到，必须 cd 进自身模块上下文）。
.PHONY: test_all
test_all: test
	go test $$(go list ./... | grep -v '/internal/') -race -count=1
	cd api && go test ./... -race -count=1

.PHONY: cover-web
# go tool cover -html coverage.txt
cover-web:
	go tool cover -html coverage.txt

.PHONY: build_linux_amd64
build_linux_amd64: | $(BIN_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o $(BIN_DIR)/app-linux-amd64 main.go

.PHONY: build_linux_arm64
build_linux_arm64: | $(BIN_DIR)
	CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS} -extldflags '-static'" -o $(BIN_DIR)/app-linux-arm64 main.go

.PHONY: build_darwin_amd64
build_darwin_amd64: | $(BIN_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o $(BIN_DIR)/app-darwin-amd64 main.go

.PHONY: build_darwin_arm64
build_darwin_arm64: | $(BIN_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o $(BIN_DIR)/app-darwin-arm64 main.go

.PHONY: build_windows
build_windows: | $(BIN_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o $(BIN_DIR)/app.exe main.go

.PHONY: ent-new
# make ent-new NAME=User
ent-new:
	go tool ent new --target internal/data/ent/schema $(NAME)

.PHONY: ent-generate
# go generate ./internal/...
ent-generate:
	go generate ./internal/data/ent/...

.PHONY: ent-hash
# db schema hash
ent-hash:
	atlas migrate hash --config file://dev/atlas.hcl --env local

.PHONY: ent-diff
# ent-diff generate diff schema sql
ent-diff:
	atlas migrate diff $(NAME) --config file://dev/atlas.hcl --env local --format '{{ sql . "  " }}'
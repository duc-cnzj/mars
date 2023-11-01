VERSION_PATH=$(shell go list -m -f "{{.Path}}")/version
LDFLAGS=-w -s  \
 -X ${VERSION_PATH}.gitBranch=$(shell git rev-parse --abbrev-ref HEAD) \
 -X ${VERSION_PATH}.buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
 -X ${VERSION_PATH}.gitCommit=$(shell git rev-parse --short HEAD) \
 -X ${VERSION_PATH}.gitTag=$(shell git describe --exact-match --tags HEAD 2> /dev/null || echo "") \
 -X ${VERSION_PATH}.kubectlVersion=$(shell go list -m -f "{{.Path}} {{.Version}}" all | grep k8s.io/client-go | cut -d " " -f2) \
 -X ${VERSION_PATH}.helmVersion=$(shell go list -m -f "{{.Path}} {{.Version}}" all | grep helm.sh/helm/v3 | cut -d " " -f2)

.PHONY: build_tools
build_tools:
	go install \
		github.com/envoyproxy/protoc-gen-validate \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		github.com/golangci/golangci-lint/cmd/golangci-lint \
		golang.org/x/tools/cmd/goimports \
		github.com/securego/gosec/v2/cmd/gosec \
		github.com/golang/mock/mockgen

.PHONY: gen_proto
gen_proto:
	cd hack && ./gen_proto.sh && cd ../ && make fmt

.PHONY: gen
gen:
	go generate ./... && make fmt

.PHONY: sec
sec:
	gosec -exclude=G104,G304 -stdout -tests=false -exclude-generated -fmt=json -out=gosec-results.json  ./...

.PHONY: lint
lint:
	golangci-lint run -D errcheck

.PHONY: release
release: build_linux_amd64 build_darwin_amd64 build_darwin_arm64

.PHONY: fmt
fmt:
	gofmt -s -w ./pkg && \
	gofmt -s -w -r 'interface{} -> any' ./internal ./plugins ./tools ./version ./third_party ./cmd && \
	goimports -w ./

.PHONY: serve
serve:
	go run -race main.go serve --debug --app_port 4000 # --grpc_port 50000

.PHONY: build_race
build_race:
	CGO_ENABLED=1 go build -ldflags="${LDFLAGS}" -race -o app main.go

.PHONY: build
build:
	CGO_ENABLED=1 go build -ldflags="${LDFLAGS}" -o app main.go

.PHONY: build_web
build_web:
	cd ./frontend && yarn build

.PHONY: test
test:
	go test ./... -race -count=1 -cover -coverprofile=cover.out

.PHONY: build_linux_amd64
build_linux_amd64:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o app-linux-amd64 main.go

.PHONY: build_linux_arm64
build_linux_arm64:
	CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS} -extldflags '-static'" -o app-linux-arm64 main.go

.PHONY: build_darwin_amd64
build_darwin_amd64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o app-darwin-amd64 main.go

.PHONY: build_darwin_arm64
build_darwin_arm64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o app-darwin-arm64 main.go

.PHONY: build_windows
build_windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o app.exe main.go
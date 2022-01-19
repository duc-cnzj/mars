VERSION_PATH=$(shell go list -m -f "{{.Path}}")/version
LDFLAGS=-w -s  \
 -X ${VERSION_PATH}.gitRepo=$(shell go list -m -f "{{.Path}}") \
 -X ${VERSION_PATH}.gitBranch=$(shell git rev-parse --abbrev-ref HEAD) \
 -X ${VERSION_PATH}.buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
 -X ${VERSION_PATH}.gitCommit=$(shell git rev-parse --short HEAD) \
 -X ${VERSION_PATH}.gitTag=$(shell git describe --exact-match --tags HEAD 2> /dev/null || echo "") \
 -X ${VERSION_PATH}.kubectlVersion=$(shell go list -m -f "{{.Path}} {{.Version}}" all | grep k8s.io/client-go | cut -d " " -f2) \
 -X ${VERSION_PATH}.helmVersion=$(shell go list -m -f "{{.Path}} {{.Version}}" all | grep helm.sh/helm/v3 | cut -d " " -f2)


.PHONY: gen
gen:
	cd hack && ./gen_proto.sh && cd .. && make fmt

.PHONY: vet
vet:
	go vet ./...

.PHONY: release
release: build_linux_amd64 build_drawin_amd64 build_drawin_arm64

.PHONY: fmt
fmt:
	gofmt -s -w ./ && goimports -w ./

.PHONY: serve
serve:
	go run main.go serve --debug --app_port 4000 --grpc_port 50000

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
	go test ./... -race -count=1 -cover

.PHONY: build_linux_amd64
build_linux_amd64:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o app-linux-amd64 main.go

.PHONY: build_linux_arm64
build_linux_arm64:
	CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS} -linkmode external -extldflags '-static'" -o app-linux-arm64 main.go

.PHONY: build_drawin_amd64
build_drawin_amd64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o app-darwin-amd64 main.go

.PHONY: build_drawin_arm64
build_drawin_arm64:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o app-darwin-arm64 main.go

.PHONY: build_windows
build_windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o app.exe main.go
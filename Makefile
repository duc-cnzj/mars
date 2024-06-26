VERSION_PATH=$(shell go list -m -f "{{.Path}}" | grep -v api)/version
PROTO_FILES=$(shell find api -name *.proto)
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
	protoc \
        --proto_path=./api \
		--proto_path ./third_party/protos \
		--go_out=paths=source_relative:./api \
		--go-grpc_out=paths=source_relative:./api \
		--openapiv2_out=./doc \
		--openapiv2_opt logtostderr=true \
		--openapiv2_opt json_names_for_fields=false \
		--grpc-gateway_out=paths=source_relative:./api \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt paths=source_relative \
		--grpc-gateway_opt generate_unbound_methods=true \
		--validate_out=lang=go,paths=source_relative:./api \
		$(PROTO_FILES)

	./frontend/node_modules/.bin/pbjs -t static-module -o ./frontend/src/api/compiled.js -w es6  ./api/**/*.proto \
      --keep-case \
      --no-verify \
      --no-convert \
      --no-create \
      --force-number \
      --force-message \
      --no-delimited
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

	./frontend/node_modules/.bin/pbts -o ./frontend/src/api/compiled.d.ts ./frontend/src/api/compiled.js --keep-case

	swagger mixin --ignore-conflicts ./third_party/doc/data/api.json ./doc/**/*.json > ./third_party/doc/data/swagger.json

.PHONY: clear_proto
clear_proto:
	rm -rf ./api/**/*.go

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
	gofmt -s -w ./api && \
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

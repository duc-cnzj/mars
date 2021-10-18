.PHONY: gen
gen:
	cd hack && ./gen_proto.sh

.PHONY: fmt
fmt:
	gofmt -w ./ && goimports -w ./

.PHONY: serve
serve:
	go run main.go grpc --debug --app_port 4000

.PHONY: build_web
build_web:
	cd ./frontend && yarn build

.PHONY: build_linux_amd64
build_linux_amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o app-linux-amd64 main.go

.PHONY: build_drawin_amd64
build_drawin_amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o app-darwin-amd64 main.go

.PHONY: build_drawin_arm64
build_drawin_arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o app-darwin-arm64 main.go

.PHONY: build_windows
build_windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o app.exe main.go
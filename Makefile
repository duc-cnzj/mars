.PHONY: fmt
fmt:
	gofmt -w ./ && goimports -w ./

.PHONY: serve
serve:
	go run main.go serve --debug --app_port 4000
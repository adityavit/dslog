.PHONY: start
start:
	go run ./cmd/server/main.go

.PHONY: generate
generate:
	protoc api/v1/*.proto --go_out=. --go_opt=paths=source_relative --proto_path=.

.PHONY: test
test:
	go test -race ./...
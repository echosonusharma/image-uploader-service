APP_BINARY=image-uploader-service

.DEFAULT_GOAL := run

build: clean
	@go build -o bin/${APP_BINARY} cmd/main.go

run: build
	@./bin/${APP_BINARY}

.PHONY: clean
clean:
	@go clean
	@rm -f bin/${APP_BINARY}

format:
	@go fmt ./...

.PHONY: test
test:
	@go test ./...



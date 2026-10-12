.PHONY: build test lint

build:
	go build -o specmg ./cmd/specmg

test:
	go test ./...

lint:
	gofmt -l . && go vet ./...

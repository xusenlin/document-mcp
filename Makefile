IMAGE := ghcr.io/xusenlin/document-mcp
VERSION := v1.0.0

.PHONY: build docker-push run

build:
	go build -o bin/document-mcp ./cmd/server/

docker-push:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		-t $(IMAGE):$(VERSION) \
		-t $(IMAGE):latest \
		--push .

run:
	go run ./cmd/server/

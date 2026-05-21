IMAGE := ghcr.io/xusenlin/document-mcp
VERSION := v1.3.1
LDFLAGS := -X main.version=$(VERSION)

.PHONY: build docker-dev docker-push run

build:
	go build -ldflags "$(LDFLAGS)" -o bin/document-mcp ./cmd/cli/

docker-dev:
	docker build --build-arg VERSION=dev -t $(IMAGE):dev .
	docker rm -f document-mcp-dev 2>/dev/null || true
	docker run -d --name document-mcp-dev -p 8080:8080 -v $(PWD):/data $(IMAGE):dev
	@echo "dev 容器已启动: document-mcp-dev (端口 8080)"

docker-push:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		-t $(IMAGE):$(VERSION) \
		-t $(IMAGE):latest \
		--push .

run:
	go run ./cmd/cli/

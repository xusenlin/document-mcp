.PHONY: build docker-build docker-run run

build:
	go build -o bin/document-mcp ./cmd/server/

docker-build:
	docker build -t document-mcp:latest .

docker-run:
	docker run -p 8080:8080 -v $(HOME)/documents:/data document-mcp:latest

run:
	go run ./cmd/server/

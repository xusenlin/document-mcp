FROM golang:1.26-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /document-mcp ./cmd/cli/

FROM debian:bookworm-slim

RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update && apt-get install -y --no-install-recommends \
        pandoc \
        libreoffice-writer \
        poppler-utils \
        python3 \
        python3-pip \
        fonts-wqy-zenhei \
        wkhtmltopdf \
        ca-certificates \
    && pip3 install --break-system-packages --no-cache-dir -i https://mirrors.aliyun.com/pypi/simple/ \
        "markitdown[docx,pdf,pptx]==0.1.5" \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /document-mcp /usr/local/bin/document-mcp

WORKDIR /data
VOLUME ["/data"]

EXPOSE 8080

ENTRYPOINT ["document-mcp"]

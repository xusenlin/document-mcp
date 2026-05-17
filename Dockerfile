FROM golang:1.26-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /document-mcp ./cmd/cli/

FROM debian:bookworm-slim

ARG TARGETARCH
ARG CHROME_VERSION=138.0.7204.183

RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update && apt-get install -y --no-install-recommends \
        pandoc \
        libreoffice-writer \
        poppler-utils \
        python3 \
        python3-pip \
        fonts-wqy-zenhei \
        libpango-1.0-0 \
        libpangocairo-1.0-0 \
        libgdk-pixbuf-2.0-0 \
        curl \
        unzip \
        ca-certificates \
        tini \
    && if [ "$TARGETARCH" = "amd64" ]; then \
        apt-get install -y --no-install-recommends \
            libasound2 \
            libatk-bridge2.0-0 \
            libcups2 \
            libdbus-1-3 \
            libdrm2 \
            libgbm1 \
            libnss3 \
            libxcomposite1 \
            libxdamage1 \
            libxext6 \
            libxfixes3 \
            libxkbcommon0 \
            libxrandr2; \
        curl -fsSL -o /tmp/headless-shell.zip "https://edgedl.me.gvt1.com/edgedl/chrome/chrome-for-testing/${CHROME_VERSION}/linux64/chrome-headless-shell-linux64.zip" \
        && unzip -q /tmp/headless-shell.zip -d /tmp/headless-shell \
        && mv /tmp/headless-shell/chrome-headless-shell-linux64/chrome-headless-shell /usr/local/bin/ \
        && chmod +x /usr/local/bin/chrome-headless-shell \
        && rm -rf /tmp/headless-shell /tmp/headless-shell.zip; \
    fi \
    && pip3 install --break-system-packages --no-cache-dir -i https://mirrors.aliyun.com/pypi/simple/ \
        "markitdown[docx,pdf,pptx]==0.1.5" \
        weasyprint \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /document-mcp /usr/local/bin/document-mcp
COPY themes/ /usr/local/share/document-mcp/themes/

WORKDIR /data
VOLUME ["/data"]

EXPOSE 8080

ENTRYPOINT ["/usr/bin/tini", "--", "document-mcp"]

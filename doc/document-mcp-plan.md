# Document MCP Server - 实施计划

## 一、项目概要

使用 Go + `github.com/modelcontextprotocol/go-sdk` 开发一个 MCP Server，通过调用本地 CLI 工具（pandoc、libreoffice、markitdown）实现文档格式转换、文本提取、PDF 合并/拆分等功能。以 Docker 容器方式部署，通过 Streamable HTTP 对外提供 MCP 服务。

## 二、当前状态

- 工作目录为空，从零构建项目。
- 已确认 MCP Go SDK 最新版为 `v1.6.0`，支持 MCP spec 2025-11-25，API 模式为 `mcp.NewServer` + `mcp.AddTool` + `server.Run(transport)`。
- Streamable HTTP 通过 `StreamableHTTPHandler` 或 `SSEHandler` 实现。

## 三、架构设计

### 3.1 文件传输方案

容器通过 Docker volume 挂载宿主机的文档目录，MCP 客户端通过 tool call 传入**容器内文件路径**。服务端直接操作挂载的文件系统，不经过 Base64 编解码，大文件也能高效处理。

```
宿主机 /home/user/docs  ──volume──>  容器内 /data
                                     ↑
                          MCP Tool: source_path="/data/report.docx"
```

### 3.2 整体架构

```
┌─────────────┐     HTTP (MCP Streamable)     ┌──────────────────┐
│  MCP Client  │ ───────────────────────────> │  document-mcp     │
│  (Claude等)  │ <─────────────────────────── │  (Go + go-sdk)    │
└─────────────┘                               │                   │
                                              │  ┌─────────────┐  │
                                              │  │ pandoc      │  │
                                              │  ├─────────────┤  │
                                              │  │ libreoffice │  │
                                              │  ├─────────────┤  │
                                              │  │ markitdown  │  │
                                              │  │ (Python)    │  │
                                              │  └─────────────┘  │
                                              └──────────────────┘
```

### 3.3 项目目录结构

```
/Users/xusenlin/Git/Github/document-mcp/
├── cmd/
│   └── server/
│       └── main.go              # 入口：创建 Server、注册 Tools、启动 HTTP
├── internal/
│   ├── converter/
│   │   ├── converter.go         # Converter 接口定义 + 公共逻辑
│   │   ├── pandoc.go            # pandoc 命令行封装
│   │   ├── libreoffice.go       # libreoffice 命令行封装
│   │   └── markitdown.go        # markitdown CLI 封装
│   ├── tool/
│   │   ├── convert.go           # 通用文档转换 tool
│   │   ├── convert_to_md.go     # 转 Markdown tool (markitdown)
│   │   ├── libreoffice.go       # LibreOffice 转换 tool
│   │   ├── formats.go           # 查询 pandoc 支持格式 tool
│   │   ├── extract_text.go      # 提取纯文本 tool
│   │   ├── merge_pdf.go         # 合并 PDF tool
│   │   └── split_pdf.go         # 拆分 PDF tool
│   └── server/
│       └── server.go            # Server 初始化 + Tool 注册
├── Dockerfile                   # 多阶段构建
├── go.mod
├── go.sum
└── Makefile
```

## 四、MCP Tools 详细设计

### Tool 1: `convert` — 通用文档格式转换（基于 pandoc）

| 字段 | 说明 |
|------|------|
| 描述 | 使用 pandoc 在不同格式之间转换文档（md↔docx↔pdf↔html↔latex 等） |
| 输入 | `source_path` (string, required): 源文件路径 |
|      | `target_path` (string, required): 目标输出路径 |
|      | `from_format` (string, optional): 源格式，不指定则自动检测 |
|      | `to_format` (string, optional): 目标格式，不指定则从 target_path 扩展名推断 |
|      | `extra_args` ([]string, optional): 额外的 pandoc 参数（如 `--toc`, `--pdf-engine=xelatex`） |
| 输出 | `success` (bool), `output_path` (string), `from_format` (string), `to_format` (string), `message` (string) |
| CLI | `pandoc <source> -f <from> -t <to> -o <target> [extra_args...]` |

### Tool 2: `convert_with_libreoffice` — LibreOffice 转换

| 字段 | 说明 |
|------|------|
| 描述 | 使用 LibreOffice 进行文档转换（如 docx→pdf 保真度更高） |
| 输入 | `source_path` (string, required): 源文件路径 |
|      | `to_format` (string, required): 目标格式（pdf, docx, odt, txt 等） |
|      | `output_dir` (string, optional): 输出目录，默认为源文件所在目录 |
| 输出 | `success` (bool), `output_path` (string), `format` (string), `message` (string) |
| CLI | `libreoffice --headless --convert-to <format> --outdir <dir> <file>` |

### Tool 3: `convert_to_markdown` — 转 Markdown（基于 markitdown）

| 字段 | 说明 |
|------|------|
| 描述 | 使用 Microsoft markitdown 将 Office 文档（docx/pptx/pdf/xlsx）转为 Markdown |
| 输入 | `source_path` (string, required): 源文件路径 |
|      | `output_path` (string, optional): 输出路径，不提供则返回 markdown 文本内容 |
| 输出 | `success` (bool), `output_path` (string, 仅指定了 output_path 时), `markdown_content` (string, 仅未指定 output_path 时), `message` (string) |
| CLI | `markitdown <source> -o <output>` 或 `markitdown <source>` (输出到 stdout) |

### Tool 4: `list_pandoc_formats` — 列出支持的格式

| 字段 | 说明 |
|------|------|
| 描述 | 列出 pandoc 支持的所有输入/输出格式，帮助 AI 了解可用的转换组合 |
| 输入 | 无 |
| 输出 | `input_formats` ([]string), `output_formats` ([]string) |
| CLI | `pandoc --list-input-formats` + `pandoc --list-output-formats` |

### Tool 5: `extract_text` — 提取纯文本

| 字段 | 说明 |
|------|------|
| 描述 | 从文档中提取纯文本内容 |
| 输入 | `source_path` (string, required): 源文件路径 |
|      | `engine` (string, optional): 提取引擎，可选 `pandoc`/`libreoffice`，默认 pandoc |
| 输出 | `success` (bool), `text` (string), `engine` (string), `page_count` (int, optional) |
| CLI | `pandoc <source> -t plain` 或 `libreoffice --headless --convert-to txt --outdir /tmp <file>` |

### Tool 6: `merge_pdf` — 合并 PDF

| 字段 | 说明 |
|------|------|
| 描述 | 将多个 PDF 文件合并为一个 PDF |
| 输入 | `source_paths` ([]string, required): 源 PDF 文件路径列表 |
|      | `output_path` (string, required): 合并后的输出路径 |
| 输出 | `success` (bool), `output_path` (string), `source_count` (int), `message` (string) |
| CLI | `pdfunite <src1> <src2> ... <output>` (poppler-utils 中的工具) |

### Tool 7: `split_pdf` — 拆分 PDF

| 字段 | 说明 |
|------|------|
| 描述 | 将 PDF 拆分为多个单页 PDF 或按页码范围拆分 |
| 输入 | `source_path` (string, required): 源 PDF 文件路径 |
|      | `output_dir` (string, required): 输出目录 |
|      | `page_range` (string, optional): 页码范围，如 "1-5,8,10-12"。不指定则每页拆一个文件 |
| 输出 | `success` (bool), `output_files` ([]string), `page_count` (int), `message` (string) |
| CLI | `pdfseparate -f <first> -l <last> <source> <output_pattern>` 或循环调用 `pdfseparate` |

## 五、Go 代码关键实现细节

### 5.1 main.go 核心结构

```go
func main() {
    server := mcp.NewServer(
        &mcp.Implementation{Name: "document-mcp", Version: "v0.1.0"},
        &mcp.ServerOptions{
            Capabilities: &mcp.ServerCapabilities{
                Tools: &mcp.ToolsCapabilities{},
            },
        },
    )

    // 注册所有 tool
    tool.RegisterAll(server)

    // 启动 Streamable HTTP
    handler := mcp.NewStreamableHTTPHandler(
        func(req *http.Request) *mcp.Server { return server },
        &mcp.StreamableHTTPOptions{},
    )
    http.ListenAndServe(":8080", handler)
}
```

### 5.2 Tool 注册模式（以 convert 为例）

```go
type ConvertInput struct {
    SourcePath string   `json:"source_path" jsonschema:"源文件路径（容器内路径）"`
    TargetPath string   `json:"target_path" jsonschema:"目标输出路径"`
    FromFormat string   `json:"from_format,omitempty" jsonschema:"源格式，不指定则自动检测"`
    ToFormat   string   `json:"to_format,omitempty" jsonschema:"目标格式，不指定从扩展名推断"`
    ExtraArgs  []string `json:"extra_args,omitempty" jsonschema:"额外 pandoc 参数"`
}

type ConvertOutput struct {
    Success    bool   `json:"success" jsonschema:"是否成功"`
    OutputPath string `json:"output_path" jsonschema:"输出文件路径"`
    FromFormat string `json:"from_format" jsonschema:"源格式"`
    ToFormat   string `json:"to_format" jsonschema:"目标格式"`
    Message    string `json:"message" jsonschema:"附加信息"`
}

func ConvertTool(ctx context.Context, req *mcp.CallToolRequest, input ConvertInput) (
    *mcp.CallToolResult, ConvertOutput, error,
) {
    // 调用 pandoc CLI
    converter := converter.NewPandoc()
    result, err := converter.Convert(ctx, input.SourcePath, input.TargetPath, ...)
    if err != nil {
        return nil, ConvertOutput{}, err
    }
    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: "转换成功: " + result.OutputPath}},
    }, ConvertOutput{...}, nil
}

// 注册
mcp.AddTool(server,
    &mcp.Tool{Name: "convert", Description: "使用 pandoc 进行文档格式转换"},
    ConvertTool,
)
```

### 5.3 Converter 接口

```go
type Converter interface {
    Convert(ctx context.Context, source, target string, opts ConvertOptions) (*ConvertResult, error)
    ListFormats(ctx context.Context) (*Formats, error)
    ExtractText(ctx context.Context, source string) (string, error)
}

type ConvertOptions struct {
    FromFormat string
    ToFormat   string
    ExtraArgs  []string
}

type ConvertResult struct {
    OutputPath string
    FromFormat string
    ToFormat   string
}
```

## 六、Dockerfile 设计

### 多阶段构建方案

```dockerfile
# ============================================
# Stage 1: 编译 Go 二进制
# ============================================
FROM golang:1.25-alpine AS builder

# 使用国内镜像加速
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /document-mcp ./cmd/server/

# ============================================
# Stage 2: 运行环境
# ============================================
FROM debian:bookworm-slim

# 配置国内镜像源
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update && apt-get install -y --no-install-recommends \
        pandoc \
        libreoffice-writer \
        libreoffice-impress \
        libreoffice-calc \
        poppler-utils \
        python3 \
        python3-pip \
        fonts-wqy-zenhei \
        ca-certificates \
    && pip3 install --no-cache-dir -i https://mirrors.aliyun.com/pypi/simple/ \
        "markitdown[docx,pdf,pptx]==0.1.5" \
    && rm -rf /var/lib/apt/lists/*

# 从 builder 阶段复制 Go 二进制
COPY --from=builder /document-mcp /usr/local/bin/document-mcp

# 默认工作目录（挂载点）
WORKDIR /data
VOLUME ["/data"]

EXPOSE 8080

ENTRYPOINT ["document-mcp"]
```

## 七、Makefile

```makefile
.PHONY: build docker-build docker-run run

# 本地编译
build:
	go build -o bin/document-mcp ./cmd/server/

# 构建 Docker 镜像
docker-build:
	docker build -t document-mcp:latest .

# 运行容器
docker-run:
	docker run -p 8080:8080 -v $(HOME)/documents:/data document-mcp:latest

# 本地运行
run:
	go run ./cmd/server/
```

## 八、关键决策与假设

| 决策 | 说明 |
|------|------|
| 传输协议 | **Streamable HTTP**，容器内运行 HTTP 服务，MCP 客户端通过 HTTP 调用 |
| 文件传输 | **文件路径方式**，通过 Docker volume 挂载，不经过 Base64 |
| 默认挂载目录 | 容器内 `/data`，用户运行时 `-v /host/path:/data` 自定义 |
| PDF 合并/拆分 | 使用 poppler-utils 中的 `pdfunite` / `pdfseparate`（比 `pdftk` 更常见） |
| markitdown 版本 | 锁定 `0.1.5`（与你给的 Dockerfile 一致） |
| Go 版本 | go.mod 中 module 为 `github.com/yourname/document-mcp` |
| 安全 | 所有 CLI 调用使用参数数组拼接，避免 shell 注入；目录遍历检测 |

## 九、验证步骤

1. `make build` — Go 编译通过
2. `make docker-build` — Docker 镜像构建成功
3. `make docker-run` — 容器启动，端口 8080 可访问
4. 使用 MCP Inspector 或 curl 测试 tool list 接口
5. 挂载测试文档目录，调用 convert tool 验证 pandoc/libreoffice/markitdown 转换成功
6. 测试 PDF 合并/拆分功能

## 十、实施步骤

1. 初始化 Go module，创建 `go.mod`
2. 创建 `cmd/server/main.go`，搭建 MCP Server + Streamable HTTP 骨架
3. 实现 `internal/converter/` 下三个转换引擎的 CLI 封装
4. 实现 `internal/tool/` 下 7 个 tool 的处理函数
5. 实现 `internal/server/server.go` 完成 tool 注册
6. 编写 `Dockerfile`（多阶段构建）
7. 编写 `Makefile`
8. 构建验证

# document-mcp

基于 Go + MCP Go SDK 的文档转换 MCP 服务，通过调用 pandoc、markitdown、LibreOffice 等本地 CLI 工具完成文档格式转换。容器化部署，Streamable HTTP 协议。

## 快速开始

```bash
# 构建镜像
make docker-build

# 运行容器（挂载宿主机文档目录）
docker run -p 8080:8080 -v /your/docs:/data ghcr.io/xusenlin/document-mcp:v1.0.0

# 推送到 ghcr.io
make docker-push
```

## MCP Tools

共 6 个 Tool，按目标格式拆分，AI 调用时只需传 `source_path`。

### 1. convert_to_markdown — 任意格式 → Markdown

| 源格式 | 引擎 | 链路 |
|--------|------|------|
| `.docx` `.pptx` `.xlsx` `.pdf` | markitdown | `src → md` |
| `.html` `.latex` `.tex` `.epub` `.odt` `.rst` `.org` `.txt` `.md` | pandoc | `src → md` |
| `.md`（同格式） | cp | 直接复制 |

### 2. convert_to_pdf — 任意格式 → PDF

| 源格式 | 引擎 | 链路 |
|--------|------|------|
| `.docx` `.pptx` `.xlsx` `.odt` | LibreOffice | `src → pdf` |
| `.md` `.html` `.latex` `.tex` `.rst` `.org` `.txt` `.epub` | pandoc | `src → pdf` |
| `.pdf`（同格式） | cp | 直接复制 |

### 3. convert_to_docx — 任意格式 → Word

| 源格式 | 引擎 | 链路 |
|--------|------|------|
| `.md` `.html` `.latex` `.tex` `.odt` `.epub` `.rst` `.org` `.txt` | pandoc | `src → docx` |
| `.pptx` `.xlsx` `.pdf` | markitdown + pandoc | `src → md → docx` |
| `.docx`（同格式） | cp | 直接复制 |

### 4. convert_to_html — 任意格式 → HTML

| 源格式 | 引擎 | 链路 |
|--------|------|------|
| `.md` `.latex` `.tex` `.docx` `.odt` `.epub` `.rst` `.org` `.txt` | pandoc | `src → html` |
| `.pptx` `.xlsx` `.pdf` | markitdown + pandoc | `src → md → html` |
| `.html`（同格式） | cp | 直接复制 |

### 5. merge_pdf — 合并 PDF

| 引擎 | 说明 |
|------|------|
| pdfunite（poppler） | 多个 PDF 合并为一个，至少 2 个文件 |

### 6. split_pdf — 拆分 PDF

| 引擎 | 说明 |
|------|------|
| pdfseparate（poppler） | 按页拆分或按页码范围拆分 |

---

## 三个引擎

| 引擎 | 角色 | 擅长 |
|------|------|------|
| **pandoc** | 通用格式转换 | 50+ 文本格式互转，覆盖面最广 |
| **markitdown** | Office → Markdown | Microsoft 出品，docx/pptx/xlsx/pdf 转 md 效果最好 |
| **LibreOffice** | Office → PDF | 排版保真，docx/pptx/xlsx → pdf 最准 |

---

## 同格式处理

当源文件与目标格式一致时（如 `.html` → `convert_to_html`），不执行转换，直接复制到输出路径。AI 无感，链路不断。

## 调用示例

```json
// md → docx（最简调用）
{
  "tool": "convert_to_docx",
  "arguments": {
    "source_path": "/data/readme.md"
  }
}
// 输出文件自动生成: /data/readme.docx
```

```json
// pptx → md（指定输出路径）
{
  "tool": "convert_to_markdown",
  "arguments": {
    "source_path": "/data/slides.pptx",
    "output_path": "/data/slides.md"
  }
}
```

```json
// 合并多个 PDF
{
  "tool": "merge_pdf",
  "arguments": {
    "source_paths": ["/data/ch1.pdf", "/data/ch2.pdf", "/data/ch3.pdf"],
    "output_path": "/data/full.pdf"
  }
}
```

## 构建

```bash
make build          # 本地编译
make docker-build   # 构建 Docker 镜像
make docker-run     # 运行容器
make run            # 本地运行
```

## 容器内工具

| 工具 | 版本 |
|------|------|
| pandoc | latest (debian bookworm) |
| libreoffice-writer | latest (debian bookworm) |
| markitdown | 0.1.5 (with docx/pdf/pptx extras) |
| pdfunite / pdfseparate | poppler-utils |
| wkhtmltopdf | 0.12.6 |

## 开发

```bash
go run ./cmd/server/                    # 本地运行（需预先安装 pandoc/markitdown/libreoffice/wkhtmltopdf）
MCP_ADDR=:9090 go run ./cmd/server/     # 指定端口
```

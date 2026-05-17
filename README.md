# document-mcp

基于 Go + MCP Go SDK 的文档转换 MCP 服务，通过调用 pandoc、markitdown、LibreOffice 等本地 CLI 工具完成文档格式转换。容器化部署，Streamable HTTP 协议。

## 快速开始

```bash

# 运行容器（挂载宿主机文档目录）
docker run -p 8080:8080 -v /data:/data ghcr.io/xusenlin/document-mcp:v1.0.1

```

## MCP Tools

共 6 个 Tool，按目标格式拆分，AI 调用时只需传 `source_path`。

### 1. convert_to_markdown — 任意格式 → Markdown

| 源格式 | 引擎 | 链路 |
|--------|------|------|
| `.docx` `.pptx` `.xlsx` `.pdf` | markitdown | `src → md` |
| `.html` `.latex` `.tex` `.epub` `.odt` `.rst` `.org` `.txt` `.md` | pandoc | `src → md` |
| `.md`（同格式） | none | 直接返回源路径 |

### 2. convert_to_pdf — 任意格式 → PDF

| 源格式 | 引擎 | 链路 |
|--------|------|------|
| `.docx` `.pptx` `.xlsx` `.odt` | LibreOffice | `src → pdf` |
| `.md` `.html` `.latex` `.tex` `.rst` `.org` `.txt` `.epub` | pandoc | `src → pdf` |
| `.pdf`（同格式） | none | 直接返回源路径 |

### 3. convert_to_docx — 任意格式 → Word

| 源格式 | 引擎 | 链路 |
|--------|------|------|
| `.md` `.html` `.latex` `.tex` `.odt` `.epub` `.rst` `.org` `.txt` | pandoc | `src → docx` |
| `.pptx` `.xlsx` `.pdf` | markitdown + pandoc | `src → md → docx` |
| `.docx`（同格式） | none | 直接返回源路径 |

### 4. convert_to_html — 任意格式 → HTML

| 源格式 | 引擎 | 链路 |
|--------|------|------|
| `.md` `.latex` `.tex` `.docx` `.odt` `.epub` `.rst` `.org` `.txt` | pandoc | `src → html` |
| `.pptx` `.xlsx` `.pdf` | markitdown + pandoc | `src → md → html` |
| `.html`（同格式） | none | 直接返回源路径 |

### 5. merge_pdf — 合并 PDF

输出固定文件名为 **merged.pdf**（在第一个源文件同目录），目标存在则报错。

| 引擎 | 说明 |
|------|------|
| pdfunite（poppler） | 多个 PDF 合并为一个，至少 2 个文件 |

### 6. split_pdf — 拆分 PDF

输出在源文件同目录，命名：按页 **{名}_page_N.pdf**，按范围 **{名}_range_N.pdf**。目标存在则报错。

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

当源文件与目标格式一致时（如 `.html` → `convert_to_html`），不执行转换，直接返回源路径。

## 输出规则

- 所有转换输出文件自动生成在**源文件同目录**，文件名与源文件相同，仅扩展名变化
- `merge_pdf` 输出固定命名为 **merged.pdf**（第一个源文件同目录）
- `split_pdf` 输出命名：按页 **{源文件名}_page_N.pdf**，按范围 **{源文件名}_range_N.pdf**
- 目标文件已存在时会**报错**，请先手动删除
- `convert_to_markdown` 支持 `return_content=true` 直接在响应中返回文本内容

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
// pptx → md 返回文本内容
{
  "tool": "convert_to_markdown",
  "arguments": {
    "source_path": "/data/slides.pptx",
    "return_content": true
  }
}
// 直接在响应中返回 markdown 文本
```

```json
// 合并多个 PDF（输出固定为 merged.pdf）
{
  "tool": "merge_pdf",
  "arguments": {
    "source_paths": ["/data/ch1.pdf", "/data/ch2.pdf", "/data/ch3.pdf"]
  }
}
// 输出: /data/merged.pdf
```

```json
// 拆分 PDF（输出命名: report_page_1.pdf, report_page_2.pdf ...）
{
  "tool": "split_pdf",
  "arguments": {
    "source_path": "/data/report.pdf"
  }
}
// 输出目录: /data/（源文件同目录）
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

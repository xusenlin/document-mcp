# document-mcp

基于 Go + MCP Go SDK 的文档转换服务。支持 **MCP** 和 **CLI + Skill** 两种模式，容器内集成 pandoc、LibreOffice、markitdown、weasyprint、headless-shell 完成任意文档格式互转。

---

## 两种使用方式

### 模式一：MCP（标准协议，通用）

通过 MCP 协议对接任意支持 MCP 的 AI 工具（OpenCode、Claude Desktop、Cursor、Windsurf 等）。

启动容器：
```bash
docker run -d -p 8080:8080 -v /your/docs:/data ghcr.io/xusenlin/document-mcp:v1.2.0
```

MCP 配置（JSON-RPC over HTTP，Streamable HTTP 传输）：
```json
{
  "mcpServers": {
    "document-mcp": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

不同工具配置位置：
- **OpenCode** — `opencode.json` 中 `mcp` 字段
- **Claude Desktop** — `claude_desktop_config.json` 中 `mcpServers` 字段
- **Cursor** — `.cursor/mcp.json`
- **Windsurf** — `mcp.json`

### 模式二：CLI + Skill（通用 AI 工具）

无需 MCP 配置，AI 通过 Skill 学习 `docker run` 命令直接调容器 CLI。本地零依赖。

```bash
# 路径规则：源文件父目录挂载到 /data，文件名不变
# /Users/xx/project/report.md  →  -v /Users/xx/project:/data  →  /data/report.md

docker run --rm -v /your/project:/data \
  ghcr.io/xusenlin/document-mcp:v1.2.0 \
  cli pdf /data/report.md
```

Skill 文件位于 `skills/document-convert/SKILL.md`，复制到对应工具的 skills 目录即可：
- Claude Desktop — `~/.claude/skills/document-convert/SKILL.md`
- OpenCode — `.opencode/skills/document-convert/SKILL.md`
- 通用 — `.agents/skills/document-convert/SKILL.md`

---

## 工具列表

### 单文件转换

| MCP Tool | CLI 命令 | 效果 |
|----------|----------|------|
| `convert_to_pdf` | `docker run --rm -v <dir>:/data <image> cli pdf /data/file.xxx` | 任意格式 → PDF |
| `convert_to_docx` | `docker run --rm -v <dir>:/data <image> cli docx /data/file.xxx` | 任意格式 → Word |
| `convert_to_html` | `docker run --rm -v <dir>:/data <image> cli html /data/file.xxx` | 任意格式 → HTML |
| `convert_to_markdown` | `docker run --rm -v <dir>:/data <image> cli markdown /data/file.xxx [return_content]` | 任意格式 → Markdown |

### PDF 操作

| MCP Tool | CLI 命令 | 效果 |
|----------|----------|------|
| `merge_pdf` | `docker run --rm -v <dir>:/data <image> cli merge /data/a.pdf /data/b.pdf` | 合并多个 PDF |
| `split_pdf` | `docker run --rm -v <dir>:/data <image> cli split /data/doc.pdf [页码范围]` | 拆分 PDF |

> `<image>` = `ghcr.io/xusenlin/document-mcp:v1.3.0`，后续命令以此为镜像。

---

## 转换引擎

### → PDF

| 源格式 | 引擎 | 链路 |
|--------|------|------|
| `.html` `.htm` | headless-shell（amd64）/ pandoc + weasyprint（arm64） | `src → pdf` |
| `.docx` `.pptx` `.xlsx` `.odt` | LibreOffice | `src → pdf` |
| `.md` `.latex` `.tex` `.rst` `.org` `.txt` `.epub` | pandoc + weasyprint | `src → pdf` |
| `.pdf` | none | 同格式跳过 |

### → Markdown

| 源格式 | 引擎 | 链路 |
|--------|------|------|
| `.docx` `.pptx` `.xlsx` `.pdf` | markitdown | `src → md` |
| `.html` `.latex` `.tex` `.epub` `.odt` `.rst` `.org` `.txt` | pandoc | `src → md` |
| `.md` | none | 同格式跳过 |

### → Word

| 源格式 | 引擎 | 链路 |
|--------|------|------|
| `.md` `.html` `.latex` `.tex` `.odt` `.epub` `.rst` `.org` `.txt` | pandoc | `src → docx` |
| `.pptx` `.xlsx` `.pdf` | markitdown + pandoc | `src → md → docx` |
| `.docx` | none | 同格式跳过 |

### → HTML

| 源格式 | 引擎 | 链路 |
|--------|------|------|
| `.md` `.latex` `.tex` `.docx` `.odt` `.epub` `.rst` `.org` `.txt` | pandoc | `src → html` |
| `.pptx` `.xlsx` `.pdf` | markitdown + pandoc | `src → md → html` |
| `.html` | none | 同格式跳过 |

---

## 输出规则

- 输出文件自动生成在**源文件同目录**，文件名相同，仅扩展名变化
- `merge_pdf` 输出固定命名为 `merged.pdf`
- `split_pdf` 输出命名为 `{名}_page_N.pdf` 或 `{名}_range_N.pdf`
- 目标文件已存在时报错，保护已有文件
- 同格式转换直接跳过，返回源路径

---

## 构建

```bash
make build          # 本地编译
make docker-dev     # 本地构建并启动测试容器
make docker-push    # 构建多架构镜像并推送
make run            # 本地运行 HTTP 模式
```

## 容器内工具

| 工具 | 版本 |
|------|------|
| pandoc | latest (debian bookworm) |
| libreoffice-writer | latest (debian bookworm) |
| markitdown | 0.1.5 (with docx/pdf/pptx extras) |
| weasyprint | latest (pip) |
| headless-shell | 138.0.7204.183 (Chromium) |
| pdfunite / pdfseparate | poppler-utils |

## 开发

```bash
go run ./cmd/cli/                        # 本地运行 HTTP 模式
MCP_ADDR=:9090 go run ./cmd/cli/         # 指定端口
go run ./cmd/cli/ pdf /data/doc.md       # CLI 模式测试
```

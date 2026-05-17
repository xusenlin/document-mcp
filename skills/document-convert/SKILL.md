---
name: document-convert
description: 文档格式转换，支持 md、pdf、docx、html、pptx、xlsx 等格式互转。通过 Docker 调用 ghcr.io/xusenlin/document-mcp 容器，本地无需安装任何转换工具。
---

# 文档转换

通过 `docker run` 调用 document-mcp 容器完成任意文档格式互转。容器内集成 pandoc、LibreOffice、markitdown、weasyprint、headless-shell。

## 路径规则

宿主机源文件父目录挂载到容器 `/data`，容器内路径 = `/data/<文件名>`：

```
宿主机: /Users/xx/project/report.md  →  -v /Users/xx/project:/data  →  /data/report.md
```

## 子命令

| 命令 | 用法 | 输出 |
|------|------|------|
| `pdf` | `docker run --rm -v <父目录>:/data ghcr.io/xusenlin/document-mcp:v1.3.1 cli pdf /data/<文件> [--theme=default\|paper]` | 同目录 .pdf |
| `docx` | `docker run --rm -v <父目录>:/data ghcr.io/xusenlin/document-mcp:v1.3.1 cli docx /data/<文件>` | 同目录 .docx |
| `html` | `docker run --rm -v <父目录>:/data ghcr.io/xusenlin/document-mcp:v1.3.1 cli html /data/<文件>` | 同目录 .html |
| `markdown` | `docker run --rm -v <父目录>:/data ghcr.io/xusenlin/document-mcp:v1.3.1 cli markdown /data/<文件>` | 同目录 .md |
| `merge` | `docker run --rm -v <父目录>:/data ghcr.io/xusenlin/document-mcp:v1.3.1 cli merge /data/a.pdf /data/b.pdf [c.pdf ...]` | 同目录 merged.pdf |
| `split` | `docker run --rm -v <父目录>:/data ghcr.io/xusenlin/document-mcp:v1.3.1 cli split /data/<文件> [页码范围]` | 同目录 {名}_page_N.pdf |

## 行为规则

- 输出文件自动生成在源文件同目录，仅扩展名变化
- 目标文件已存在时会报错，需先手动删除后再试
- 同格式转换（如 pdf→pdf）会跳过并返回源路径
- `merge` 输出固定命名为 `merged.pdf`
- `split` 输出命名为 `{名}_page_N.pdf`，指定范围时 `{名}_range_N.pdf`
- `pdf` 支持 `--theme=` 参数：`default`（GitHub 技术文档风格）/ `paper`（学术报告风格），仅对 md/latex/tex/rst/org/txt/epub 转 PDF 生效

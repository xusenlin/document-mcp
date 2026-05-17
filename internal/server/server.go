package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xusenlin/document-mcp/internal/tool"
)

func RegisterTools(s *mcp.Server) {
	mcp.AddTool(s,
		&mcp.Tool{
			Name: "convert_to_markdown",
			Description: `将任意文档格式转换为 Markdown。

输出规则：
- 输出文件自动生成在源文件同目录，文件名与源文件相同、扩展名为 .md
- return_content=true：直接在响应中返回 markdown 文本内容
- return_content=false（默认）：存盘并返回输出路径
- 源文件已是 .md 格式时，return_content=true 直接读取返回，return_content=false 会报错

支持的输入格式：
- Office 文档（走 markitdown）: docx, pptx, xlsx, pdf
- 文本格式（走 pandoc）: html, latex, tex, epub, odt, rst, org, txt, md`,
		},
		tool.ConvertToMarkdown,
	)

	mcp.AddTool(s,
		&mcp.Tool{
			Name: "convert_to_pdf",
			Description: `将任意文档格式转换为 PDF。

输出规则：
- 输出文件自动生成在源文件同目录，文件名与源文件相同、扩展名为 .pdf
- 源文件已是 .pdf 格式时，直接返回源路径，无需转换
- 目标文件已存在时会报错，请先手动删除

Office 文档（docx/pptx/xlsx/odt）走 LibreOffice 渲染，排版保真度最高。
文本格式（md/html/latex/tex/rst/org/txt/epub）走 pandoc 转换。`,
		},
		tool.ConvertToPDF,
	)

	mcp.AddTool(s,
		&mcp.Tool{
			Name: "convert_to_docx",
			Description: `将任意文档格式转换为 Word docx。

输出规则：
- 输出文件自动生成在源文件同目录，文件名与源文件相同、扩展名为 .docx
- 源文件已是 .docx 格式时，直接返回源路径，无需转换
- 目标文件已存在时会报错，请先手动删除

支持直接转换（走 pandoc）: md, html, latex, tex, odt, epub, rst, org, txt
支持多步转换（自动走 markitdown→md→pandoc 链）: pptx, xlsx, pdf`,
		},
		tool.ConvertToDocx,
	)

	mcp.AddTool(s,
		&mcp.Tool{
			Name: "convert_to_html",
			Description: `将任意文档格式转换为 HTML。

输出规则：
- 输出文件自动生成在源文件同目录，文件名与源文件相同、扩展名为 .html
- 源文件已是 .html 格式时，直接返回源路径，无需转换
- 目标文件已存在时会报错，请先手动删除

支持直接转换（走 pandoc）: md, latex, tex, docx, odt, epub, rst, org, txt, html
支持多步转换（自动走 markitdown→md→pandoc 链）: pptx, xlsx, pdf`,
		},
		tool.ConvertToHTML,
	)

	mcp.AddTool(s,
		&mcp.Tool{
			Name: "merge_pdf",
			Description: `将多个 PDF 文件合并为一个 PDF 文件。

输出规则：
- 输出文件自动生成在第一个源文件同目录，固定文件名为 merged.pdf
- 目标文件已存在时会报错，请先手动删除
- 所有输入文件必须是 pdf 格式，至少需要 2 个文件`,
		},
		tool.MergePDF,
	)

	mcp.AddTool(s,
		&mcp.Tool{
			Name: "split_pdf",
			Description: `将 PDF 文件拆分为多个 PDF 文件。

输出规则：
- 拆分后的文件生成在源文件同目录
- 按页拆分：文件名为 {源文件名}_page_N.pdf
- 按范围拆分：文件名为 {源文件名}_range_N.pdf
- 任一输出文件已存在时会报错，请先手动删除

支持两种模式：
1. 不指定 page_range：按页拆分，每页一个 PDF 文件
2. 指定 page_range：按页码范围拆分，如 "1-5" 或 "1-3,6,8-10"`,
		},
		tool.SplitPDF,
	)
}

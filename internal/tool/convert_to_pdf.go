package tool

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xusenlin/document-mcp/internal/converter"
)

type ConvertToPDFInput struct {
	SourcePath string `json:"source_path" jsonschema:"源文件路径（容器内路径）"`
	Theme      string `json:"theme,omitempty" jsonschema:"PDF 主题：default（GitHub 技术文档风格）或 paper（学术报告风格），默认 default。仅对 md/latex/tex/rst/org/txt/epub 转 PDF 生效"`
}

type ConvertToPDFOutput struct {
	Success    bool   `json:"success" jsonschema:"是否成功"`
	OutputPath string `json:"output_path" jsonschema:"输出文件路径（在源文件同目录）"`
	Engine     string `json:"engine" jsonschema:"使用的转换引擎"`
	Chain      string `json:"chain" jsonschema:"转换链路"`
	Message    string `json:"message" jsonschema:"附加信息"`
}

func ConvertToPDF(ctx context.Context, req *mcp.CallToolRequest, input ConvertToPDFInput) (
	*mcp.CallToolResult, ConvertToPDFOutput, error,
) {
	sourceExt := converter.Ext(input.SourcePath)

	if converter.IsSameFormat(input.SourcePath, "pdf") {
		msg := fmt.Sprintf("✅ 源文件已是 PDF 格式，直接返回\n• 输出: %s", input.SourcePath)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}},
			ConvertToPDFOutput{Success: true, OutputPath: input.SourcePath, Engine: "none",
				Chain: "同格式，无需转换", Message: "source is already pdf"}, nil
	}

	targetPath := resolveOutput(input.SourcePath, "pdf")

	if err := checkTargetExists(targetPath); err != nil {
		return nil, ConvertToPDFOutput{}, err
	}

	var res *converter.ConvertResult
	var err error

	if converter.IsLibreofficeExt(sourceExt) {
		res, err = converter.NewLibreOffice().ConvertToPDF(ctx, input.SourcePath, filepath.Dir(targetPath))
		if err == nil {
			if filepath.Base(res.OutputPath) != filepath.Base(targetPath) {
				converter.CopyFile(res.OutputPath, targetPath)
				res.OutputPath = targetPath
			}
		}
	} else if sourceExt == "html" || sourceExt == "htm" {
		headless := converter.NewHeadlessShell()
		if headless.Available() {
			res, err = headless.ConvertToPDF(ctx, input.SourcePath, targetPath)
		} else {
			p := newPandoc(input.Theme)
			res, err = p.Convert(ctx, input.SourcePath, targetPath)
		}
	} else if converter.IsPandocInputExt(sourceExt) {
		p := newPandoc(input.Theme)
		res, err = p.Convert(ctx, input.SourcePath, targetPath)
	} else {
		return nil, ConvertToPDFOutput{},
			fmt.Errorf("unsupported source format .%s for pdf conversion, supported Office: docx/pptx/xlsx/odt, text: md/html/latex/tex/rst/org/txt/epub", sourceExt)
	}

	if err != nil {
		return nil, ConvertToPDFOutput{}, err
	}

	msg := formatResult(res)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}},
		ConvertToPDFOutput{Success: true, OutputPath: res.OutputPath, Engine: res.Engine,
			Chain: res.Chain, Message: "conversion successful"}, nil
}

func newPandoc(theme string) *converter.Pandoc {
	p := converter.NewPandoc()
	if theme != "" {
		p.PdfTheme = theme
	}
	return p
}

package tool

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xusenlin/document-mcp/internal/converter"
)

type ConvertToHTMLInput struct {
	SourcePath string `json:"source_path" jsonschema:"源文件路径（容器内路径）"`
	OutputPath string `json:"output_path,omitempty" jsonschema:"输出文件路径，不指定则自动生成"`
}

type ConvertToHTMLOutput struct {
	Success    bool   `json:"success" jsonschema:"是否成功"`
	OutputPath string `json:"output_path" jsonschema:"输出文件路径"`
	Engine     string `json:"engine" jsonschema:"使用的转换引擎"`
	Chain      string `json:"chain" jsonschema:"转换链路"`
	Message    string `json:"message" jsonschema:"附加信息"`
}

func ConvertToHTML(ctx context.Context, req *mcp.CallToolRequest, input ConvertToHTMLInput) (
	*mcp.CallToolResult, ConvertToHTMLOutput, error,
) {
	sourceExt := converter.Ext(input.SourcePath)

	if converter.IsSameFormat(input.SourcePath, "html") {
		targetPath := resolveOutput(input.SourcePath, input.OutputPath, "html")
		nopRes, err := converter.NopCopy(input.SourcePath, targetPath)
		if err != nil {
			return nil, ConvertToHTMLOutput{}, err
		}
		msg := formatResult(nopRes)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}},
			ConvertToHTMLOutput{Success: true, OutputPath: targetPath, Engine: "copy", Chain: nopRes.Chain,
				Message: "same format, copied directly"}, nil
	}

	targetPath := resolveOutput(input.SourcePath, input.OutputPath, "html")

	if converter.RequiresMultiStep(sourceExt, "html") {
		res, err := converter.MultiStep(ctx, input.SourcePath, targetPath, "html")
		if err != nil {
			return nil, ConvertToHTMLOutput{}, err
		}
		msg := formatResult(res)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}},
			ConvertToHTMLOutput{Success: true, OutputPath: res.OutputPath, Engine: res.Engine,
				Chain: res.Chain, Message: "conversion successful"}, nil
	}

	if !converter.IsPandocInputExt(sourceExt) && sourceExt != "docx" {
		return nil, ConvertToHTMLOutput{},
			fmt.Errorf("unsupported source format .%s for html conversion, supported: %s, docx (pptx/xlsx/pdf via multi-step)", sourceExt, pandocSupported)
	}

	res, err := converter.NewPandoc().Convert(ctx, input.SourcePath, targetPath)
	if err != nil {
		return nil, ConvertToHTMLOutput{}, err
	}

	msg := formatResult(res)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: msg}}},
		ConvertToHTMLOutput{Success: true, OutputPath: res.OutputPath, Engine: res.Engine,
			Chain: res.Chain, Message: "conversion successful"}, nil
}

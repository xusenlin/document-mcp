package tool

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xusenlin/document-mcp/internal/converter"
)

type ConvertToMarkdownInput struct {
	SourcePath string `json:"source_path" jsonschema:"源文件路径（容器内路径）"`
	OutputPath string `json:"output_path,omitempty" jsonschema:"输出文件路径，不指定则直接在 content 中返回 markdown 文本"`
}

type ConvertToMarkdownOutput struct {
	Success    bool   `json:"success" jsonschema:"是否成功"`
	OutputPath string `json:"output_path,omitempty" jsonschema:"输出文件路径"`
	Engine     string `json:"engine" jsonschema:"使用的转换引擎"`
	Chain      string `json:"chain" jsonschema:"转换链路"`
	Content    string `json:"content,omitempty" jsonschema:"转换后的 Markdown 文本（未指定 output_path 时）"`
	Message    string `json:"message" jsonschema:"附加信息"`
}

func ConvertToMarkdown(ctx context.Context, req *mcp.CallToolRequest, input ConvertToMarkdownInput) (
	*mcp.CallToolResult, ConvertToMarkdownOutput, error,
) {
	sourceExt := converter.Ext(input.SourcePath)

	if converter.IsSameFormat(input.SourcePath, "md") {
		targetPath := resolveOutput(input.SourcePath, input.OutputPath, "md")
		nopRes, err := converter.NopCopy(input.SourcePath, targetPath)
		if err != nil {
			return nil, ConvertToMarkdownOutput{}, err
		}
		msg := formatResult(nopRes)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, ConvertToMarkdownOutput{
			Success: true, OutputPath: targetPath, Engine: "copy",
			Chain: nopRes.Chain, Message: "same format, copied directly",
		}, nil
	}

	targetPath := resolveOutput(input.SourcePath, input.OutputPath, "md")

	var res *converter.ConvertResult
	var err error

	if converter.IsMarkitdownExt(sourceExt) {
		res, err = converter.NewMarkitdown().Convert(ctx, input.SourcePath, targetPath)
	} else if converter.IsPandocInputExt(sourceExt) {
		res, err = converter.NewPandoc().Convert(ctx, input.SourcePath, targetPath)
	} else {
		return nil, ConvertToMarkdownOutput{},
			fmt.Errorf("unsupported source format .%s, supported: %s, %s", sourceExt, markitdownSupported, pandocSupported)
	}

	if err != nil {
		return nil, ConvertToMarkdownOutput{}, err
	}

	out := ConvertToMarkdownOutput{
		Success: true, OutputPath: res.OutputPath, Engine: res.Engine,
		Chain: res.Chain, Message: "conversion successful",
	}

	if input.OutputPath == "" {
		data, readErr := converter.ReadFile(res.OutputPath)
		if readErr != nil {
			return nil, ConvertToMarkdownOutput{}, readErr
		}
		out.Content = string(data)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: out.Content}},
		}, out, nil
	}

	msg := formatResult(res)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, out, nil
}

func formatResult(res *converter.ConvertResult) string {
	return fmt.Sprintf("✅ 转换完成\n• 输出: %s\n• 引擎: %s\n• 链路: %s",
		res.OutputPath, res.Engine, res.Chain)
}

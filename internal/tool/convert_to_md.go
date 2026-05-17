package tool

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/xusenlin/document-mcp/internal/converter"
)

type ConvertToMarkdownInput struct {
	SourcePath    string `json:"source_path" jsonschema:"源文件路径（容器内路径）"`
	ReturnContent bool   `json:"return_content,omitempty" jsonschema:"是否直接在响应中返回 markdown 文本内容（不存盘），默认 false 即存盘"`
}

type ConvertToMarkdownOutput struct {
	Success    bool   `json:"success" jsonschema:"是否成功"`
	OutputPath string `json:"output_path,omitempty" jsonschema:"输出文件路径（存盘模式下在源文件同目录）"`
	Engine     string `json:"engine" jsonschema:"使用的转换引擎"`
	Chain      string `json:"chain" jsonschema:"转换链路"`
	Content    string `json:"content,omitempty" jsonschema:"转换后的 Markdown 文本（内容返回模式下）"`
	Message    string `json:"message" jsonschema:"附加信息"`
}

func ConvertToMarkdown(ctx context.Context, req *mcp.CallToolRequest, input ConvertToMarkdownInput) (
	*mcp.CallToolResult, ConvertToMarkdownOutput, error,
) {
	sourceExt := converter.Ext(input.SourcePath)

	if converter.IsSameFormat(input.SourcePath, "md") {
		if input.ReturnContent {
			data, readErr := converter.ReadFile(input.SourcePath)
			if readErr != nil {
				return nil, ConvertToMarkdownOutput{}, readErr
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
			}, ConvertToMarkdownOutput{
				Success: true, OutputPath: input.SourcePath, Engine: "none",
				Chain: "同格式，直接读取", Content: string(data),
				Message: "source is already markdown, content returned",
			}, nil
		}
		return nil, ConvertToMarkdownOutput{},
			fmt.Errorf("源文件已是 Markdown 格式，无法再存盘。请使用 return_content=true 直接读取")
	}

	targetPath := resolveOutput(input.SourcePath, "md")

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

	if input.ReturnContent {
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


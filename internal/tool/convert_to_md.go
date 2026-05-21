package tool

import (
	"context"
	"fmt"
	"os"

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
			result := &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
			}
			out := ConvertToMarkdownOutput{
				Success: true, OutputPath: input.SourcePath, Engine: "none",
				Chain: "同格式，直接读取", Content: string(data),
				Message: "source is already markdown, content returned",
			}
			return result, out, nil
		}
		return nil, ConvertToMarkdownOutput{},
			fmt.Errorf("源文件已是 Markdown 格式，无法再存盘。请使用 return_content=true 直接读取")
	}

	if input.ReturnContent {
		return convertToMarkdownContent(ctx, input.SourcePath, sourceExt)
	}

	targetPath := resolveOutput(input.SourcePath, "md")

	if err := checkTargetExists(targetPath); err != nil {
		return nil, ConvertToMarkdownOutput{}, err
	}

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

	msg := formatResult(res)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, out, nil
}

func convertToMarkdownContent(ctx context.Context, sourcePath, sourceExt string) (
	*mcp.CallToolResult, ConvertToMarkdownOutput, error,
) {
	var res *converter.ConvertResult
	var err error

	if converter.IsMarkitdownExt(sourceExt) {
		res, err = converter.NewMarkitdown().Convert(ctx, sourcePath, "-")
	} else if converter.IsPandocInputExt(sourceExt) {
		res, err = convertPandocToTempMarkdown(ctx, sourcePath)
	} else {
		return nil, ConvertToMarkdownOutput{},
			fmt.Errorf("unsupported source format .%s, supported: %s, %s", sourceExt, markitdownSupported, pandocSupported)
	}
	if err != nil {
		return nil, ConvertToMarkdownOutput{}, err
	}
	defer os.Remove(res.OutputPath)

	data, readErr := converter.ReadFile(res.OutputPath)
	if readErr != nil {
		return nil, ConvertToMarkdownOutput{}, readErr
	}

	content := string(data)
	result := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: content}},
	}
	out := ConvertToMarkdownOutput{
		Success: true, Engine: res.Engine, Chain: res.Chain,
		Content: content, Message: "conversion successful, content returned",
	}
	return result, out, nil
}

func convertPandocToTempMarkdown(ctx context.Context, sourcePath string) (*converter.ConvertResult, error) {
	tmpFile, err := os.CreateTemp("", "_mcp_md_*.md")
	if err != nil {
		return nil, fmt.Errorf("create temp markdown: %w", err)
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("close temp markdown: %w", err)
	}

	res, err := converter.NewPandoc().Convert(ctx, sourcePath, tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return nil, err
	}
	return res, nil
}

package tool

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MergePDFInput struct {
	SourcePaths []string `json:"source_paths" jsonschema:"源 PDF 文件路径列表（至少 2 个），第一个源文件的目录将作为输出目录"`
}

type MergePDFOutput struct {
	Success     bool   `json:"success" jsonschema:"是否成功"`
	OutputPath  string `json:"output_path" jsonschema:"合并后的文件路径（在第一个源文件同目录，文件名为 merged.pdf）"`
	SourceCount int    `json:"source_count" jsonschema:"合并的源文件数量"`
	Message     string `json:"message" jsonschema:"附加信息"`
}

func MergePDF(ctx context.Context, req *mcp.CallToolRequest, input MergePDFInput) (
	*mcp.CallToolResult, MergePDFOutput, error,
) {
	if len(input.SourcePaths) < 2 {
		return nil, MergePDFOutput{}, fmt.Errorf("at least 2 source files required")
	}

	for _, p := range input.SourcePaths {
		if !strings.EqualFold(filepath.Ext(p), ".pdf") {
			return nil, MergePDFOutput{}, fmt.Errorf("not a pdf file: %s", p)
		}
	}

	outputPath := filepath.Join(filepath.Dir(input.SourcePaths[0]), "merged.pdf")

	if err := checkTargetExists(outputPath); err != nil {
		return nil, MergePDFOutput{}, err
	}

	args := append(input.SourcePaths, outputPath)
	cmd := exec.CommandContext(ctx, "pdfunite", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, MergePDFOutput{}, fmt.Errorf("pdfunite failed: %s: %w", string(output), err)
	}

	text := fmt.Sprintf("✅ PDF 合并完成\n• 输出: %s\n• 合并了 %d 个文件\n• 输出位置: 第一个源文件同目录下的 merged.pdf",
		outputPath, len(input.SourcePaths))

	return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		},
		MergePDFOutput{
			Success: true, OutputPath: outputPath,
			SourceCount: len(input.SourcePaths), Message: "merged successfully",
		}, nil
}

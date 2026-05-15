package tool

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SplitPDFInput struct {
	SourcePath string `json:"source_path" jsonschema:"源 PDF 文件路径"`
	OutputDir  string `json:"output_dir" jsonschema:"输出目录"`
	PageRange  string `json:"page_range,omitempty" jsonschema:"页码范围，如 1-5 或 1-3,6,8-10。不指定则每页拆一个文件"`
}

type SplitPDFOutput struct {
	Success     bool     `json:"success" jsonschema:"是否成功"`
	OutputFiles []string `json:"output_files" jsonschema:"拆分的输出文件列表"`
	PageCount   int      `json:"page_count" jsonschema:"拆分的页数"`
	Message     string   `json:"message" jsonschema:"附加信息"`
}

func SplitPDF(ctx context.Context, req *mcp.CallToolRequest, input SplitPDFInput) (
	*mcp.CallToolResult, SplitPDFOutput, error,
) {
	baseName := strings.TrimSuffix(filepath.Base(input.SourcePath), ".pdf")

	if input.PageRange == "" {
		pattern := filepath.Join(input.OutputDir, baseName+"_page_%d.pdf")
		cmd := exec.CommandContext(ctx, "pdfseparate", input.SourcePath, pattern)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, SplitPDFOutput{}, fmt.Errorf("pdfseparate failed: %s: %w", string(output), err)
		}

		pageInfo, err := exec.CommandContext(ctx, "pdfinfo", input.SourcePath).CombinedOutput()
		if err != nil {
			return nil, SplitPDFOutput{}, err
		}
		pageCount := 0
		for _, line := range strings.Split(string(pageInfo), "\n") {
			if strings.HasPrefix(line, "Pages:") {
				fmt.Sscanf(line, "Pages: %d", &pageCount)
				break
			}
		}

		var files []string
		for i := 1; i <= pageCount; i++ {
			files = append(files, fmt.Sprintf(filepath.Join(input.OutputDir, baseName+"_page_%d.pdf"), i))
		}

		text := fmt.Sprintf("✅ PDF 拆分完成\n• 输出目录: %s\n• 共 %d 页",
			input.OutputDir, pageCount)

		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}},
			SplitPDFOutput{Success: true, OutputFiles: files, PageCount: pageCount,
				Message: "split by page"}, nil
	}

	ranges := strings.Split(input.PageRange, ",")
	var outputFiles []string
	pageCount := 0

	for i, r := range ranges {
		r = strings.TrimSpace(r)
		parts := strings.Split(r, "-")
		var first, last int

		if len(parts) == 2 {
			first, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
			last, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		} else {
			first, _ = strconv.Atoi(parts[0])
			last = first
		}

		outFile := filepath.Join(input.OutputDir, fmt.Sprintf("%s_range_%d.pdf", baseName, i+1))
		pattern := filepath.Join(input.OutputDir, fmt.Sprintf("%s_range_%d_%%d.pdf", baseName, i+1))

		cmd := exec.CommandContext(ctx, "pdfseparate", "-f", strconv.Itoa(first), "-l", strconv.Itoa(last),
			input.SourcePath, pattern)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, SplitPDFOutput{}, fmt.Errorf("pdfseparate failed: %s: %w", string(output), err)
		}

		pageFiles := pageSlice(pattern, first, last)
		if len(pageFiles) == 1 {
			if err := renameFile(pageFiles[0], outFile); err != nil {
				return nil, SplitPDFOutput{}, err
			}
		} else {
			mergeArgs := append(pageFiles, outFile)
			if err := exec.CommandContext(ctx, "pdfunite", mergeArgs...).Run(); err != nil {
				return nil, SplitPDFOutput{}, err
			}
			for _, pf := range pageFiles {
				exec.Command("rm", pf).Run()
			}
		}
		outputFiles = append(outputFiles, outFile)
		pageCount += (last - first + 1)
	}

	text := fmt.Sprintf("✅ PDF 按范围拆分完成\n• 输出目录: %s\n• 生成 %d 个文件，共 %d 页",
		input.OutputDir, len(outputFiles), pageCount)

	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}},
		SplitPDFOutput{Success: true, OutputFiles: outputFiles, PageCount: pageCount,
			Message: "split by range"}, nil
}

func pageSlice(pattern string, first, last int) []string {
	var files []string
	for i := first; i <= last; i++ {
		files = append(files, fmt.Sprintf(pattern, i))
	}
	return files
}

func renameFile(src, dst string) error {
	cmd := exec.Command("mv", src, dst)
	return cmd.Run()
}

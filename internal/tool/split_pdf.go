package tool

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SplitPDFInput struct {
	SourcePath string `json:"source_path" jsonschema:"源 PDF 文件路径，拆分后的文件将生成在源文件同目录"`
	PageRange  string `json:"page_range,omitempty" jsonschema:"页码范围，如 1-5 或 1-3,6,8-10。不指定则每页拆一个文件"`
}

type SplitPDFOutput struct {
	Success     bool     `json:"success" jsonschema:"是否成功"`
	OutputFiles []string `json:"output_files" jsonschema:"拆分后的输出文件路径列表（在源文件同目录）"`
	PageCount   int      `json:"page_count" jsonschema:"拆分的页数"`
	Message     string   `json:"message" jsonschema:"附加信息"`
}

func SplitPDF(ctx context.Context, req *mcp.CallToolRequest, input SplitPDFInput) (
	*mcp.CallToolResult, SplitPDFOutput, error,
) {
	if !strings.EqualFold(filepath.Ext(input.SourcePath), ".pdf") {
		return nil, SplitPDFOutput{}, fmt.Errorf("not a pdf file: %s", input.SourcePath)
	}

	baseName := strings.TrimSuffix(filepath.Base(input.SourcePath), filepath.Ext(input.SourcePath))
	outputDir := filepath.Dir(input.SourcePath)
	totalPages, err := getPDFPageCount(ctx, input.SourcePath)
	if err != nil {
		return nil, SplitPDFOutput{}, err
	}

	if input.PageRange == "" {
		pattern := filepath.Join(outputDir, baseName+"_page_%d.pdf")

		for i := 1; i <= totalPages; i++ {
			path := fmt.Sprintf(pattern, i)
			if err := checkTargetExists(path); err != nil {
				return nil, SplitPDFOutput{}, err
			}
		}

		cmd := exec.CommandContext(ctx, "pdfseparate", input.SourcePath, pattern)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, SplitPDFOutput{}, fmt.Errorf("pdfseparate failed: %s: %w", string(output), err)
		}

		var files []string
		for i := 1; i <= totalPages; i++ {
			files = append(files, fmt.Sprintf(pattern, i))
		}

		text := fmt.Sprintf("✅ PDF 拆分完成\n• 输出目录: %s（源文件同目录）\n• 共 %d 页\n• 命名: %s_page_N.pdf",
			outputDir, totalPages, baseName)

		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}},
			SplitPDFOutput{Success: true, OutputFiles: files, PageCount: totalPages,
				Message: "split by page"}, nil
	}

	ranges, err := parsePageRanges(input.PageRange, totalPages)
	if err != nil {
		return nil, SplitPDFOutput{}, err
	}
	for i := range ranges {
		outFile := filepath.Join(outputDir, fmt.Sprintf("%s_range_%d.pdf", baseName, i+1))
		if err := checkTargetExists(outFile); err != nil {
			return nil, SplitPDFOutput{}, err
		}
	}

	tmpDir, err := os.MkdirTemp(outputDir, "."+baseName+"_split_*")
	if err != nil {
		return nil, SplitPDFOutput{}, fmt.Errorf("create temp split dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	var outputFiles []string
	pageCount := 0

	for i, r := range ranges {
		outFile := filepath.Join(outputDir, fmt.Sprintf("%s_range_%d.pdf", baseName, i+1))

		pattern := filepath.Join(tmpDir, fmt.Sprintf("%s_range_%d_%%d.pdf", baseName, i+1))
		cmd := exec.CommandContext(ctx, "pdfseparate", "-f", strconv.Itoa(r.first), "-l", strconv.Itoa(r.last),
			input.SourcePath, pattern)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, SplitPDFOutput{}, fmt.Errorf("pdfseparate failed: %s: %w", string(output), err)
		}

		pageFiles := pageSlice(pattern, r.first, r.last)
		if len(pageFiles) == 1 {
			if err := os.Rename(pageFiles[0], outFile); err != nil {
				return nil, SplitPDFOutput{}, err
			}
		} else {
			mergeArgs := append(pageFiles, outFile)
			output, err := exec.CommandContext(ctx, "pdfunite", mergeArgs...).CombinedOutput()
			if err != nil {
				return nil, SplitPDFOutput{}, fmt.Errorf("pdfunite failed: %s: %w", string(output), err)
			}
		}

		outputFiles = append(outputFiles, outFile)
		pageCount += r.last - r.first + 1
	}

	text := fmt.Sprintf("✅ PDF 按范围拆分完成\n• 输出目录: %s（源文件同目录）\n• 生成 %d 个文件，共 %d 页\n• 命名: %s_range_N.pdf",
		outputDir, len(outputFiles), pageCount, baseName)

	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}},
		SplitPDFOutput{Success: true, OutputFiles: outputFiles, PageCount: pageCount,
			Message: "split by range"}, nil
}

type pageRange struct {
	first int
	last  int
}

func getPDFPageCount(ctx context.Context, sourcePath string) (int, error) {
	pageInfo, err := exec.CommandContext(ctx, "pdfinfo", sourcePath).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("pdfinfo failed: %s: %w", string(pageInfo), err)
	}
	for _, line := range strings.Split(string(pageInfo), "\n") {
		if strings.HasPrefix(line, "Pages:") {
			pageCount, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Pages:")))
			if err != nil {
				return 0, fmt.Errorf("parse pdf page count: %w", err)
			}
			if pageCount < 1 {
				return 0, fmt.Errorf("pdf has no pages: %s", sourcePath)
			}
			return pageCount, nil
		}
	}
	return 0, fmt.Errorf("pdfinfo output missing page count: %s", sourcePath)
}

func parsePageRanges(raw string, totalPages int) ([]pageRange, error) {
	parts := strings.Split(raw, ",")
	ranges := make([]pageRange, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("invalid page range %q: empty segment", raw)
		}

		bounds := strings.Split(part, "-")
		if len(bounds) > 2 {
			return nil, fmt.Errorf("invalid page range segment %q", part)
		}

		first, err := parsePositivePage(bounds[0], part)
		if err != nil {
			return nil, err
		}
		last := first
		if len(bounds) == 2 {
			last, err = parsePositivePage(bounds[1], part)
			if err != nil {
				return nil, err
			}
		}

		if last < first {
			return nil, fmt.Errorf("invalid page range segment %q: end page is before start page", part)
		}
		if last > totalPages {
			return nil, fmt.Errorf("invalid page range segment %q: page %d exceeds total pages %d", part, last, totalPages)
		}

		ranges = append(ranges, pageRange{first: first, last: last})
	}

	return ranges, nil
}

func parsePositivePage(raw, segment string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("invalid page range segment %q: missing page number", segment)
	}
	page, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid page range segment %q: %w", segment, err)
	}
	if page < 1 {
		return 0, fmt.Errorf("invalid page range segment %q: page must be >= 1", segment)
	}
	return page, nil
}

func pageSlice(pattern string, first, last int) []string {
	var files []string
	for i := first; i <= last; i++ {
		files = append(files, fmt.Sprintf(pattern, i))
	}
	return files
}

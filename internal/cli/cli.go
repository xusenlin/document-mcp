package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/xusenlin/document-mcp/internal/tool"
)

const usage = `document-mcp cli — 文档转换命令行

用法:
  document-mcp cli pdf       <源文件> [--theme=default|paper]
  document-mcp cli docx      <源文件>
  document-mcp cli html      <源文件>
  document-mcp cli markdown  <源文件> [return_content]
  document-mcp cli merge     <源文件1> <源文件2> [源文件3 ...]
  document-mcp cli split     <源文件> [页码范围]`

func Run(args []string) {
	if len(args) > 0 && args[0] == "cli" {
		args = args[1:]
	}
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	ctx := context.Background()
	cmd := args[0]
	tail := args[1:]

	switch cmd {
	case "pdf":
		if len(tail) < 1 {
			fail("用法: document-mcp cli pdf <源文件> [--theme=default|paper]")
		}
		theme := parseTheme(tail[1:])
		_, out, err := tool.ConvertToPDF(ctx, nil, tool.ConvertToPDFInput{SourcePath: tail[0], Theme: theme})
		printOrFail(out.OutputPath, err)

	case "docx":
		if len(tail) < 1 {
			fail("用法: document-mcp cli docx <源文件>")
		}
		_, out, err := tool.ConvertToDocx(ctx, nil, tool.ConvertToDocxInput{SourcePath: tail[0]})
		printOrFail(out.OutputPath, err)

	case "html":
		if len(tail) < 1 {
			fail("用法: document-mcp cli html <源文件>")
		}
		_, out, err := tool.ConvertToHTML(ctx, nil, tool.ConvertToHTMLInput{SourcePath: tail[0]})
		printOrFail(out.OutputPath, err)

	case "markdown":
		if len(tail) < 1 {
			fail("用法: document-mcp cli markdown <源文件> [return_content]")
		}
		returnContent := len(tail) > 1 && tail[1] == "return_content"
		_, out, err := tool.ConvertToMarkdown(ctx, nil, tool.ConvertToMarkdownInput{
			SourcePath:    tail[0],
			ReturnContent: returnContent,
		})
		if err != nil {
			fail(err.Error())
		}
		if returnContent && out.Content != "" {
			fmt.Println(out.Content)
		} else {
			fmt.Println(out.OutputPath)
		}

	case "merge":
		if len(tail) < 2 {
			fail("用法: document-mcp cli merge <源文件1> <源文件2> [源文件3 ...]")
		}
		_, out, err := tool.MergePDF(ctx, nil, tool.MergePDFInput{SourcePaths: tail})
		printOrFail(out.OutputPath, err)

	case "split":
		if len(tail) < 1 {
			fail("用法: document-mcp cli split <源文件> [页码范围]")
		}
		pageRange := ""
		if len(tail) > 1 {
			pageRange = tail[1]
		}
		_, out, err := tool.SplitPDF(ctx, nil, tool.SplitPDFInput{
			SourcePath: tail[0],
			PageRange:  pageRange,
		})
		if err != nil {
			fail(err.Error())
		}
		for _, f := range out.OutputFiles {
			fmt.Println(f)
		}

	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n\n%s\n", cmd, usage)
		os.Exit(1)
	}
}

func printOrFail(outputPath string, err error) {
	if err != nil {
		fail(err.Error())
	}
	fmt.Println(outputPath)
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "Error:", msg)
	os.Exit(1)
}

func parseTheme(args []string) string {
	for _, a := range args {
		if strings.HasPrefix(a, "--theme=") {
			return strings.TrimPrefix(a, "--theme=")
		}
	}
	return ""
}

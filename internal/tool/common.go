package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xusenlin/document-mcp/internal/converter"
)

const (
	markitdownSupported = "docx, pptx, xlsx, pdf"
	pandocSupported     = "md, html, latex, tex, epub, odt, rst, org, txt"
)

func resolveOutput(sourcePath, targetExt string) string {
	dir := filepath.Dir(sourcePath)
	base := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	return filepath.Join(dir, base+"."+targetExt)
}

func formatResult(res *converter.ConvertResult) string {
	return fmt.Sprintf("✅ 转换完成\n• 输出: %s\n• 引擎: %s\n• 链路: %s",
		res.OutputPath, res.Engine, res.Chain)
}

func checkTargetExists(targetPath string) error {
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("目标文件已存在: %s，请手动删除后再试", targetPath)
	}
	return nil
}

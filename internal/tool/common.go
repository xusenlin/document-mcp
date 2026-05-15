package tool

import (
	"path/filepath"
	"strings"
)

const (
	markitdownSupported = "docx, pptx, xlsx, pdf"
	pandocSupported     = "md, html, latex, tex, epub, odt, rst, org, txt"
)

func resolveOutput(sourcePath, outputPath, defaultExt string) string {
	if outputPath != "" {
		return outputPath
	}
	dir := filepath.Dir(sourcePath)
	base := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	return filepath.Join(dir, base+"."+defaultExt)
}

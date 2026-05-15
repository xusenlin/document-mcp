package converter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrSameFormat = fmt.Errorf("same format, no conversion needed")

type ConvertResult struct {
	OutputPath string
	SourceExt  string
	TargetExt  string
	Engine     string
	Chain      string
}

func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return fmt.Errorf("write target: %w", err)
	}
	return nil
}

func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func NopCopy(src, dst string) (*ConvertResult, error) {
	if err := CopyFile(src, dst); err != nil {
		return nil, err
	}
	srcExt := strings.TrimPrefix(filepath.Ext(src), ".")
	tgtExt := strings.TrimPrefix(filepath.Ext(dst), ".")
	return &ConvertResult{
		OutputPath: dst,
		SourceExt:  srcExt,
		TargetExt:  tgtExt,
		Engine:     "copy",
		Chain:      "same format, copied directly",
	}, nil
}

func Ext(path string) string {
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
}

func IsMarkitdownExt(ext string) bool {
	e := strings.ToLower(ext)
	return e == "pdf" || e == "docx" || e == "pptx" || e == "xlsx"
}

func IsPandocInputExt(ext string) bool {
	e := strings.ToLower(ext)
	switch e {
	case "md", "markdown", "html", "htm", "latex", "tex", "epub", "odt", "rst", "org", "txt",
		"textile", "wiki", "mediawiki", "csv", "tsv":
		return true
	}
	return false
}

func IsLibreofficeExt(ext string) bool {
	e := strings.ToLower(ext)
	switch e {
	case "docx", "pptx", "xlsx", "odt", "ods", "odp":
		return true
	}
	return false
}

func IsOfficeExt(ext string) bool {
	e := strings.ToLower(ext)
	return e == "docx" || e == "pptx" || e == "xlsx"
}

func IsSameFormat(path, ext string) bool {
	srcExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	tgtExt := strings.ToLower(strings.TrimPrefix(ext, "."))
	return srcExt == tgtExt
}

func RequiresMultiStep(sourceExt, targetExt string) bool {
	se := strings.ToLower(sourceExt)
	te := strings.ToLower(targetExt)
	return (se == "pptx" || se == "xlsx" || se == "pdf") && (te == "docx" || te == "html")
}

func MultiStep(ctx context.Context, sourcePath, targetPath, targetExt string) (*ConvertResult, error) {
	tmpDir := os.TempDir()
	baseName := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	tmpMd := filepath.Join(tmpDir, fmt.Sprintf("_mcp_tmp_%s.md", baseName))
	defer os.Remove(tmpMd)

	md := NewMarkitdown()
	mdRes, err := md.Convert(ctx, sourcePath, tmpMd)
	if err != nil {
		return nil, fmt.Errorf("step1 markitdown: %w", err)
	}

	p := NewPandoc()
	panRes, err := p.Convert(ctx, tmpMd, targetPath)
	if err != nil {
		return nil, fmt.Errorf("step2 pandoc: %w", err)
	}

	return &ConvertResult{
		OutputPath: targetPath,
		SourceExt:  mdRes.SourceExt,
		TargetExt:  panRes.TargetExt,
		Engine:     "markitdown+pandoc",
		Chain:      fmt.Sprintf("%s → md → %s", mdRes.SourceExt, panRes.TargetExt),
	}, nil
}

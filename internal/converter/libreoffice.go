package converter

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type LibreOffice struct{}

func NewLibreOffice() *LibreOffice {
	return &LibreOffice{}
}

func (l *LibreOffice) ConvertToPDF(ctx context.Context, sourcePath, outputDir string) (*ConvertResult, error) {
	sourceExt := strings.TrimPrefix(filepath.Ext(sourcePath), ".")

	if outputDir == "" {
		outputDir = filepath.Dir(sourcePath)
	}

	args := []string{
		"--headless",
		"--convert-to", "pdf",
		"--outdir", outputDir,
		sourcePath,
	}

	cmd := exec.CommandContext(ctx, "libreoffice", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("libreoffice failed: %s: %w", string(output), err)
	}

	baseName := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	outputPath := filepath.Join(outputDir, baseName+".pdf")

	return &ConvertResult{
		OutputPath: outputPath,
		SourceExt:  sourceExt,
		TargetExt:  "pdf",
		Engine:     "libreoffice",
		Chain:      fmt.Sprintf("%s → pdf", sourceExt),
	}, nil
}

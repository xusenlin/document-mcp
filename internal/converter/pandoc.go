package converter

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type Pandoc struct{}

func NewPandoc() *Pandoc {
	return &Pandoc{}
}

func (p *Pandoc) Convert(ctx context.Context, sourcePath, targetPath string) (*ConvertResult, error) {
	sourceExt := strings.TrimPrefix(filepath.Ext(sourcePath), ".")
	targetExt := strings.TrimPrefix(filepath.Ext(targetPath), ".")

	args := []string{
		sourcePath,
		"-o", targetPath,
		"--wrap=none",
	}

	if targetExt == "pdf" {
		args = append(args, "--pdf-engine=weasyprint")
	}
	if targetExt == "html" {
		args = append(args, "--standalone")
	}

	cmd := exec.CommandContext(ctx, "pandoc", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pandoc failed: %s: %w", string(output), err)
	}

	return &ConvertResult{
		OutputPath: targetPath,
		SourceExt:  sourceExt,
		TargetExt:  targetExt,
		Engine:     "pandoc",
		Chain:      fmt.Sprintf("%s → %s", sourceExt, targetExt),
	}, nil
}

func (p *Pandoc) ExtractText(ctx context.Context, sourcePath string) (string, error) {
	cmd := exec.CommandContext(ctx, "pandoc", sourcePath, "-t", "plain", "--wrap=none")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pandoc extract text failed: %s: %w", string(output), err)
	}
	return string(output), nil
}

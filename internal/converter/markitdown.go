package converter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Markitdown struct{}

func NewMarkitdown() *Markitdown {
	return &Markitdown{}
}

func (m *Markitdown) Convert(ctx context.Context, sourcePath, targetPath string) (*ConvertResult, error) {
	sourceExt := strings.TrimPrefix(filepath.Ext(sourcePath), ".")
	targetExt := strings.TrimPrefix(filepath.Ext(targetPath), ".")

	if targetPath == "" || targetPath == "-" {
		return m.convertToStdout(ctx, sourcePath)
	}

	args := []string{sourcePath, "-o", targetPath}
	cmd := exec.CommandContext(ctx, "markitdown", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("markitdown failed: %s: %w", string(output), err)
	}

	return &ConvertResult{
		OutputPath: targetPath,
		SourceExt:  sourceExt,
		TargetExt:  targetExt,
		Engine:     "markitdown",
		Chain:      fmt.Sprintf("%s → %s", sourceExt, targetExt),
	}, nil
}

func (m *Markitdown) convertToStdout(ctx context.Context, sourcePath string) (*ConvertResult, error) {
	cmd := exec.CommandContext(ctx, "markitdown", sourcePath)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("markitdown failed: %s: %w", string(data), err)
	}

	sourceExt := strings.TrimPrefix(filepath.Ext(sourcePath), ".")
	tmpFile, err := os.CreateTemp("", "_mcp_md_*.md")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("close temp file: %w", err)
	}

	return &ConvertResult{
		OutputPath: tmpPath,
		SourceExt:  sourceExt,
		TargetExt:  "md",
		Engine:     "markitdown",
		Chain:      fmt.Sprintf("%s → md", sourceExt),
	}, nil
}

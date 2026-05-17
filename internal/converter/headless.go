package converter

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type HeadlessShell struct{}

func NewHeadlessShell() *HeadlessShell {
	return &HeadlessShell{}
}

func (h *HeadlessShell) Available() bool {
	_, err := exec.LookPath("chrome-headless-shell")
	return err == nil
}

func (h *HeadlessShell) ConvertToPDF(ctx context.Context, sourcePath, targetPath string) (*ConvertResult, error) {
	if !h.Available() {
		return nil, fmt.Errorf("headless-shell not available on this platform")
	}

	cmd := exec.CommandContext(ctx, "chrome-headless-shell",
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--print-to-pdf="+targetPath,
		"--no-pdf-header-footer",
		"file://"+sourcePath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("headless-shell failed: %s: %w", string(output), err)
	}

	sourceExt := strings.TrimPrefix(filepath.Ext(sourcePath), ".")
	return &ConvertResult{
		OutputPath: targetPath,
		SourceExt:  sourceExt,
		TargetExt:  "pdf",
		Engine:     "headless-shell",
		Chain:      fmt.Sprintf("%s → pdf", sourceExt),
	}, nil
}

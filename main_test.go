package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessTemplate_StripsCommentsAndStatus(t *testing.T) {
	in := "start\n<!-- hidden\nline -->\nvisible\n@status 7\n"
	out, code := processTemplate(in)

	if code != 7 {
		t.Fatalf("expected status 7, got %d", code)
	}
	if strings.Contains(out, "hidden") || strings.Contains(out, "@status") {
		t.Fatalf("unexpected output: %q", out)
	}
	if !strings.Contains(out, "visible") {
		t.Fatalf("expected visible text in output: %q", out)
	}
}

func TestProcessTemplate_InvalidStatusIsNotPrinted(t *testing.T) {
	in := "text\n@status abc\n"
	out, code := processTemplate(in)
	if code != 0 {
		t.Fatalf("expected default status 0, got %d", code)
	}
	if strings.Contains(out, "@status") {
		t.Fatalf("status directive should not be printed: %q", out)
	}
}

func TestProcessTemplate_UnclosedCommentIsPreserved(t *testing.T) {
	in := "text\n<!-- broken\nnext"
	out, code := processTemplate(in)
	if code != 0 {
		t.Fatalf("expected default status 0, got %d", code)
	}
	if !strings.Contains(out, "<!-- broken") {
		t.Fatalf("expected malformed comment to be preserved: %q", out)
	}
}

func TestProcessTemplate_RemovesSurroundingBlankLines(t *testing.T) {
	out, _ := processTemplate("\n\nline\n\n")
	if strings.HasPrefix(out, "\n") || strings.HasSuffix(out, "\n") {
		t.Fatalf("expected no leading/trailing newlines: %q", out)
	}
	if out != "line" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestProcessTemplate_CollapsesDuplicateBlankLines(t *testing.T) {
	in := "a\n\n\n\nb\n\n\nc"
	out, _ := processTemplate(in)
	if strings.Contains(out, "\n\n\n") {
		t.Fatalf("expected no multiple blank lines: %q", out)
	}
	if !strings.Contains(out, "a\n\nb") || !strings.Contains(out, "b\n\nc") {
		t.Fatalf("expected single blank lines between paragraphs: %q", out)
	}
}

func TestRunInit_DefaultPathFromTempDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	t.Setenv("TMP", tmp)
	t.Setenv("TEMP", tmp)

	if err := runInit(nil, ioDiscard{}, ioDiscard{}); err != nil {
		t.Fatalf("runInit failed: %v", err)
	}

	path := filepath.Join(tmp, defaultFileName)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist at %s: %v", path, err)
	}
}

func TestRunInit_RespectsForceFlag(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "custom.md")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	if err := runInit([]string{path}, ioDiscard{}, ioDiscard{}); err == nil {
		t.Fatal("expected error when file exists without force")
	}

	if err := runInit([]string{"--force", path}, ioDiscard{}, ioDiscard{}); err != nil {
		t.Fatalf("expected force overwrite to succeed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read overwritten file: %v", err)
	}
	if !strings.Contains(string(data), "@status 0") {
		t.Fatalf("unexpected template content: %q", string(data))
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

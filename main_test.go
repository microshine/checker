package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessTemplate_StripsCommentsAndStatus(t *testing.T) {
	in := "start\n<!-- hidden\nline -->\nvisible\n@status 7\n"
	stdout, stderr, code := processTemplate(in)

	if code != 7 {
		t.Fatalf("expected status 7, got %d", code)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr output, got: %q", stderr)
	}
	if strings.Contains(stdout, "hidden") || strings.Contains(stdout, "@status") {
		t.Fatalf("unexpected output: %q", stdout)
	}
	if !strings.Contains(stdout, "visible") {
		t.Fatalf("expected visible text in output: %q", stdout)
	}
}

func TestProcessTemplate_InvalidStatusIsNotPrinted(t *testing.T) {
	in := "text\n@status abc\n"
	stdout, stderr, code := processTemplate(in)
	if code != 0 {
		t.Fatalf("expected default status 0, got %d", code)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr output, got: %q", stderr)
	}
	if strings.Contains(stdout, "@status") {
		t.Fatalf("status directive should not be printed: %q", stdout)
	}
}

func TestProcessTemplate_UnclosedCommentIsPreserved(t *testing.T) {
	in := "text\n<!-- broken\nnext"
	stdout, stderr, code := processTemplate(in)
	if code != 0 {
		t.Fatalf("expected default status 0, got %d", code)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr output, got: %q", stderr)
	}
	if !strings.Contains(stdout, "<!-- broken") {
		t.Fatalf("expected malformed comment to be preserved: %q", stdout)
	}
}

func TestProcessTemplate_RemovesSurroundingBlankLines(t *testing.T) {
	stdout, stderr, _ := processTemplate("\n\nline\n\n")
	if stderr != "" {
		t.Fatalf("expected no stderr output, got: %q", stderr)
	}
	if strings.HasPrefix(stdout, "\n") || strings.HasSuffix(stdout, "\n") {
		t.Fatalf("expected no leading/trailing newlines: %q", stdout)
	}
	if stdout != "line" {
		t.Fatalf("unexpected output: %q", stdout)
	}
}

func TestProcessTemplate_CollapsesDuplicateBlankLines(t *testing.T) {
	in := "a\n\n\n\nb\n\n\nc"
	stdout, stderr, _ := processTemplate(in)
	if stderr != "" {
		t.Fatalf("expected no stderr output, got: %q", stderr)
	}
	if strings.Contains(stdout, "\n\n\n") {
		t.Fatalf("expected no multiple blank lines: %q", stdout)
	}
	if !strings.Contains(stdout, "a\n\nb") || !strings.Contains(stdout, "b\n\nc") {
		t.Fatalf("expected single blank lines between paragraphs: %q", stdout)
	}
}

func TestProcessTemplate_RoutesTextToStderr(t *testing.T) {
	in := "a\n@stderr\nerr\n@stdout\nb\n"
	stdout, stderr, code := processTemplate(in)
	if code != 0 {
		t.Fatalf("expected status 0, got %d", code)
	}
	if stdout != "a\nb" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if stderr != "err" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestProcessTemplate_RepeatsMarkers(t *testing.T) {
	in := "@stderr\nerr1\n@stdout\nok\n@stderr\nerr2\n"
	stdout, stderr, code := processTemplate(in)
	if code != 0 {
		t.Fatalf("expected status 0, got %d", code)
	}
	if stdout != "ok" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if stderr != "err1\nerr2" {
		t.Fatalf("unexpected stderr: %q", stderr)
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

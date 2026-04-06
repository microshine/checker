package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const defaultFileName = ".check.md"
const version = "1.2.0"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		switch args[0] {
		case "help", "-h", "--help":
			printHelp(stdout)
			return 0
		case "version", "-v", "--version":
			_, _ = fmt.Fprintln(stdout, version)
			return 0
		case "init":
			if err := runInit(args[1:], stdout, stderr); err != nil {
				_, _ = fmt.Fprintln(stderr, err)
				return 1
			}
			return 0
		default:
			exitCode, out, errOut, err := runTemplateMode()
			if err != nil {
				_, _ = fmt.Fprintln(stderr, err)
				return 1
			}
			if out != "" {
				writeOutput(stdout, out)
			}
			if errOut != "" {
				writeOutput(stderr, errOut)
			}
			return exitCode
		}
	}

	exitCode, out, errOut, err := runTemplateMode()
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	if out != "" {
		writeOutput(stdout, out)
	}
	if errOut != "" {
		writeOutput(stderr, errOut)
	}
	return exitCode
}

func writeOutput(w io.Writer, out string) {
	_, _ = fmt.Fprint(w, out)
	if !strings.HasSuffix(out, "\n") {
		_, _ = fmt.Fprintln(w)
	}
}

func runInit(args []string, stdout, _ io.Writer) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	force := fs.Bool("force", false, "overwrite existing file")
	fs.BoolVar(force, "f", false, "overwrite existing file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	path := defaultTemplatePath()
	if fs.NArg() > 1 {
		return errors.New("only one path argument is supported")
	}
	if fs.NArg() == 1 {
		path = fs.Arg(0)
	}

	content := "<!-- checker template example:\n" +
		"@status 0           # define exit status\n" +
		"@stdout            # following text will be written to stdout\n" +
		"This text goes to stdout.\n" +
		"@stderr            # following text will be written to stderr\n" +
		"This text goes to stderr.\n" +
		"@stdout            # back to stdout\n" +
		"More stdout text.\n" +
		"-->\n" +
		"@status 0\n" +
		"# check\n\n" +
		"This is a simple stdout message.\n"
	if err := writeTemplateFile(path, content, *force); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(stdout, path)
	return nil
}

func runTemplateMode() (int, string, string, error) {
	path := os.Getenv("CHECK_FILE")
	if path == "" {
		path = defaultTemplatePath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to read template %q: %w", path, err)
	}

	stdoutOutput, stderrOutput, status := processTemplate(string(data))
	return status, stdoutOutput, stderrOutput, nil
}

func printHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "checker - process markdown template and return exit status")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  checker                Run template mode")
	_, _ = fmt.Fprintln(w, "  checker help           Show help")
	_, _ = fmt.Fprintln(w, "  checker version        Show version")
	_, _ = fmt.Fprintln(w, "  checker init [flags] [path]")
	_, _ = fmt.Fprintln(w, "    -f, --force        Overwrite existing file in init mode")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Environment:")
	_, _ = fmt.Fprintln(w, "  CHECK_FILE           Path to template file for template mode")
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Template format:")
	_, _ = fmt.Fprintln(w, "  - Markdown file treated as plain text")
	_, _ = fmt.Fprintln(w, "  - HTML comments <!-- ... --> are ignored")
	_, _ = fmt.Fprintln(w, "  - Optional directives: @status <code>, @stdout, @stderr")
}

func defaultTemplatePath() string {
	return filepath.Join(os.TempDir(), defaultFileName)
}

func writeTemplateFile(path, content string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("file already exists: %s", path)
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

func processTemplate(input string) (string, string, int) {
	// Remove UTF-8 BOM if present
	input = strings.TrimPrefix(input, "\xef\xbb\xbf")
	withoutComments := stripHTMLComments(input)

	status := 0
	stdoutLines := []string{}
	stderrLines := []string{}
	current := &stdoutLines

	scanner := bufio.NewScanner(strings.NewReader(withoutComments))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "@status") {
			parts := strings.Fields(trimmed)
			if len(parts) == 2 {
				if code, err := strconv.Atoi(parts[1]); err == nil {
					status = code
				}
			}
			continue
		}
		switch trimmed {
		case "@stdout":
			current = &stdoutLines
			continue
		case "@stderr":
			current = &stderrLines
			continue
		}
		*current = append(*current, line)
	}

	return normalizeOutput(stdoutLines), normalizeOutput(stderrLines), status
}

func normalizeOutput(lines []string) string {
	// Trim leading and trailing blank lines
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines) - 1
	for end >= start && strings.TrimSpace(lines[end]) == "" {
		end--
	}
	if start > end {
		return ""
	}
	filtered := lines[start : end+1]

	// Collapse consecutive blank lines to a single blank line
	var resultLines []string
	prevBlank := false
	for _, l := range filtered {
		isBlank := strings.TrimSpace(l) == ""
		if isBlank {
			if prevBlank {
				continue
			}
			prevBlank = true
			resultLines = append(resultLines, "")
			continue
		}
		prevBlank = false
		resultLines = append(resultLines, l)
	}

	return strings.Join(resultLines, "\n")
}

func stripHTMLComments(s string) string {
	for {
		start := strings.Index(s, "<!--")
		if start == -1 {
			return s
		}
		end := strings.Index(s[start+4:], "-->")
		if end == -1 {
			return s
		}
		s = s[:start] + s[start+4+end+3:]
	}
}

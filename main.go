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
const version = "1.1.0"

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
			exitCode, output, err := runTemplateMode()
			if err != nil {
				_, _ = fmt.Fprintln(stderr, err)
				return 1
			}
			_, _ = fmt.Fprint(stdout, output)
			return exitCode
		}
	}

	exitCode, output, err := runTemplateMode()
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	_, _ = fmt.Fprint(stdout, output)
	return exitCode
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
	_, _ = fmt.Fprintln(w, "  - Optional directive: @status <code>")
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

	content := "<!-- check template -->\n# check\n\n@status 0\n"
	if err := writeTemplateFile(path, content, *force); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(stdout, path)
	return nil
}

func runTemplateMode() (int, string, error) {
	path := os.Getenv("CHECK_FILE")
	if path == "" {
		path = defaultTemplatePath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read template %q: %w", path, err)
	}

	output, status := processTemplate(string(data))
	return status, output, nil
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

func processTemplate(input string) (string, int) {
	// Remove UTF-8 BOM if present
	input = strings.TrimPrefix(input, "\xef\xbb\xbf")
	withoutComments := stripHTMLComments(input)
	hadTrailingNewline := strings.HasSuffix(withoutComments, "\n")
	status := 0
	var outLines []string
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
		outLines = append(outLines, line)
	}

	output := strings.Join(outLines, "\n")
	if hadTrailingNewline {
		output += "\n"
	}
	return output, status
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

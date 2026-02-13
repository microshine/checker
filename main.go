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

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		switch args[0] {
		case "help", "-h", "--help":
			printHelp(stdout)
			return 0
		case "init":
			if err := runInit(args[1:], stdout, stderr); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			return 0
		default:
			exitCode, output, err := runTemplateMode()
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
			fmt.Fprint(stdout, output)
			return exitCode
		}
	}

	exitCode, output, err := runTemplateMode()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprint(stdout, output)
	return exitCode
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "check - process markdown template and return exit status")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  check                Run template mode")
	fmt.Fprintln(w, "  check help           Show help")
	fmt.Fprintln(w, "  check init [flags] [path]")
	fmt.Fprintln(w, "    -f, --force        Overwrite existing file in init mode")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Environment:")
	fmt.Fprintln(w, "  CHECK_FILE           Path to template file for template mode")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Template format:")
	fmt.Fprintln(w, "  - Markdown file treated as plain text")
	fmt.Fprintln(w, "  - HTML comments <!-- ... --> are ignored")
	fmt.Fprintln(w, "  - Optional directive: @status <code>")
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

	fmt.Fprintln(stdout, path)
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
	if strings.HasPrefix(input, "\xef\xbb\xbf") {
		input = input[3:]
	}
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

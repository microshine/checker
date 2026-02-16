# checker

A simple CLI tool for processing markdown templates and returning exit status codes. It can be used for checklists, status reports, or any templated output with conditional exit codes.

## Features

- Process markdown files as templates
- Strip HTML comments (`<!-- -->`)
- Set exit status via `@status <code>` directive
- Template mode for dynamic output
- Init mode to create default templates

## Installation

### Global Installation

To install the tool globally:

```bash
go install github.com/microshine/checker@latest
```

For the latest development version:

```bash
go install github.com/microshine/checker@main
```

### Local Build

Clone the repository and build locally:

```bash
git clone https://github.com/microshine/checker.git
cd checker
go build -o checker .
```

## Usage

### Template Mode

Run template mode to process a template file:

```bash
checker [path]
```

By default, it looks for `.check.md` in the temp directory, or set `CHECK_FILE` environment variable, or specify path as argument.

### Init Mode

Create a new template file:

```bash
checker init [path]
```

Options:

- `-f, --force`: Overwrite existing file

### Help

```bash
checker help
```

## Template Format

- Markdown content is output as-is
- HTML comments are stripped
- Use `@status <number>` to set exit code (e.g., `@status 1`)

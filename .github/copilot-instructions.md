# VS Code Copilot Commit Message Guidelines

This repository follows a clear Conventional Commits-style format for commit messages.
Use these rules when generating commit summaries or when proposing a Git commit message.

## Required format

- Use a short, imperative sentence in the subject line.
- Start with one of the allowed types:
  - `feat:` for new features
  - `fix:` for bug fixes
  - `docs:` for documentation changes
  - `chore:` for maintenance tasks
  - `ci:` for continuous integration changes
  - `test:` for test-related changes
  - `release:` for version bump or release tagging
- Keep the subject line concise, ideally under 72 characters.
- Do not use a trailing period in the subject.
- Separate the subject and body with a blank line when additional explanation is needed.

## Preferred style

- Write the subject in lowercase after the type prefix, e.g. `fix: handle missing newline in output`.
- Use imperative mood: `add`, `fix`, `update`, `remove`, `apply`.
- If a body is included, explain why the change is needed and any important details.

## Example commit messages

- `feat: add @stdout/@stderr output zones and update init template`
- `fix: enhance template processing by removing surrounding blank lines`
- `docs: clarify usage instructions for template mode in README`
- `chore: add newline for better readability in README.md`
- `release: v1.1.1`

## Purpose for Copilot

When generating a commit message, follow the repository pattern exactly.
If the change is minor and self-explanatory, keep only the subject.
If the change is more complex, add a short body separated by a blank line.

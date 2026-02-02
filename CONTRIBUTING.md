# Contributing to File Deduplicator

Thank you for your interest in contributing to File Deduplicator! This document provides guidelines for contributing to the project.

## Table of Contents

- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Testing](#testing)
- [Commit Conventions](#commit-conventions)
- [Pull Request Process](#pull-request-process)
- [Bug Reporting](#bug-reporting)
- [Feature Requests](#feature-requests)

## Development Setup

### Prerequisites

- Go 1.16 or higher
- Git
- Make (optional, for building with Makefile)

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/luinbytes/file-deduplicator.git
cd file-deduplicator

# Build from source
go build -o file-deduplicator main.go

# Or use Makefile
make build
```

### Run Tests

```bash
# Run all tests
go test -v

# Run tests with coverage
go test -v -cover

# Or use Makefile
make test
```

### Development Workflow

1. Create a new branch for your feature or fix:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. Make your changes
3. Run tests to ensure nothing is broken
4. Commit your changes (see [Commit Conventions](#commit-conventions))
5. Push to your fork or branch
6. Create a pull request (see [Pull Request Process](#pull-request-process))

## Code Style

### Go Conventions

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` to format your code:
  ```bash
  gofmt -w main.go
  ```

- Use `go vet` to check for common mistakes:
  ```bash
  go vet ./...
  ```

### Naming Conventions

- **Package names**: lowercase, single word, short
- **Exported functions**: PascalCase
- **Private functions**: camelCase
- **Constants**: PascalCase or UPPER_SNAKE_CASE
- **Variables**: camelCase

### Comments

- Document exported functions, types, and constants
- Use `go doc` to view documentation:
  ```bash
  go deduplicator
  ```

- Keep comments clear and concise
- Explain *why*, not *what* (code should be self-explanatory)

### Error Handling

- Always check for errors
- Provide context with error messages:
  ```go
  return fmt.Errorf("failed to read file %s: %w", path, err)
  ```

- Use `errors.Is` and `errors.As` for error checking

### Concurrency

- Use goroutines for parallel processing
- Use channels for communication
- Use `sync.WaitGroup` for waiting
- Be careful with shared state

## Testing

### Write Tests

- Write tests for new functionality
- Use table-driven tests where appropriate:
  ```go
  tests := []struct {
      name  string
      input string
      want  string
  }{
      {"test case 1", "input", "expected"},
      {"test case 2", "input2", "expected2"},
  }

  for _, tt := range tests {
      t.Run(tt.name, func(t *testing.T) {
          // test implementation
      })
  }
  ```

### Test Coverage

- Aim for >80% code coverage
- Use race detector for concurrent code:
  ```bash
  go test -race ./...
  ```

### Testing Edge Cases

- Test empty directories
- Test with hidden files
- Test with permission errors
- Test with different file sizes
- Test with different hash algorithms

## Commit Conventions

### Commit Message Format

```
<type>: <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, no logic changes)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Build process, tooling, dependencies

### Examples

```
feat: add support for BLAKE3 hash algorithm

Implement BLAKE3 hash algorithm option with benchmarks
showing 20% performance improvement over SHA256.

Fixes #42
```

```
fix: handle permission denied errors gracefully

Added error handling for directories the user doesn't have
read access to. Previously, the tool would crash with a
panic.

Closes #15
```

### Footer

- Reference issues with `Fixes #42` or `Closes #15`
- Use `BREAKING CHANGE:` for breaking changes

## Pull Request Process

### Before Submitting

1. Ensure your code follows the [code style](#code-style)
2. Run tests: `go test -v`
3. Format code: `gofmt -w .`
4. Check for issues: `go vet ./...`
5. Update documentation if needed

### PR Description Template

```markdown
## Description
Brief description of what this PR does.

## Changes
- Bullet point 1
- Bullet point 2
- Bullet point 3

## Testing
How you tested this change.

## Related Issues
Fixes #42
Closes #15
```

### Review Process

1. Automated tests must pass
2. Code review required
3. At least one approval needed
4. PR must be up-to-date with main branch

### After Merge

- Delete your branch (optional)
- Consider updating documentation

## Bug Reporting

### Before Reporting

- Check if bug is already reported
- Check if there's an existing fix in a recent version
- Verify you're using the latest version

### Bug Report Template

Use the [GitHub issue form](../../issues/new?template=bug_report.md)

Include:

- Go version
- Operating system and version
- File deduplicator version
- Steps to reproduce
- Expected behavior
- Actual behavior
- Any error messages

## Feature Requests

### Guidelines

- Check if feature is already requested
- Describe the use case clearly
- Consider if it fits the project scope
- Be open to discussion

### Feature Request Template

Use the [GitHub issue form](../../issues/new?template=feature_request.md)

Include:

- Feature description
- Why you need this feature
- How you envision it working
- Possible alternatives

## Getting Help

- Open a GitHub issue for bugs or questions
- Check existing documentation and issues first
- Be respectful and patient with maintainers

---

Thank you for contributing to File Deduplicator! 🚀

# Contributing to github-notifier

Thanks for your interest in contributing! This document outlines the process for
contributing to this project.

## Quick Start

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Commit using conventional commits
6. Push and open a Pull Request

## Development Setup

```bash
git clone https://github.com/your-fork/github-notifier
cd github-notifier
make deps    # Install system dependencies (Arch)
go mod download
make test   # Verify everything works
```

## Code Style

- Run `go fmt` before committing
- Keep functions small and focused
- Add tests for new functionality
- Document exported functions with Go doc comments

## Testing

Run the full test suite:

```bash
make test
```

The project uses Go's built-in testing package. Test files follow the naming
convention `*_test.go`.

## Commit Messages

We use conventional commits:

```
<type>(<scope>): <description>

[optional body]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code refactoring
- `test`: Adding/updating tests
- `docs`: Documentation changes
- `chore`: Maintenance tasks

Example:
```
feat(tray): add dynamic icon color based on notification count
```

## Pull Request Process

1. Ensure all tests pass
2. Update documentation if needed
3. PRs should target the `master` branch
4. Describe your changes in the PR description

## Reporting Bugs

Use the bug report template and include:
- Go version (`go version`)
- OS and distro
- Steps to reproduce
- Relevant log output

## Feature Requests

Open an issue with:
- Clear description of the feature
- Use case (why do you need it?)
- Potential implementation approaches (optional)

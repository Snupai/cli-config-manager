# Contributing to dotman

Thank you for your interest in contributing to dotman! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please read it before contributing.

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in the Issues section
2. Use the bug report template when creating a new issue
3. Include as much detail as possible:
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Environment details (OS, version, etc.)
   - Screenshots if applicable

### Suggesting Features

1. Check if the feature has already been suggested
2. Use the feature request template
3. Provide a clear description of the feature
4. Explain why this feature would be useful
5. Include any relevant examples or use cases

### Pull Requests

1. Fork the repository
2. Create a new branch for your feature/fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. Make your changes
4. Run tests:
   ```bash
   go test ./...
   ```
5. Commit your changes with clear commit messages
6. Push to your fork
7. Create a Pull Request

### Development Setup

1. Install Go (1.21 or later)
2. Clone the repository:
   ```bash
   git clone https://github.com/Snupai/cli-config-manager.git
   cd cli-config-manager
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Build the project:
   ```bash
   go build
   ```

### Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Write clear and concise comments
- Keep functions small and focused
- Use meaningful variable and function names

### Testing

- Write tests for new features
- Ensure all tests pass before submitting a PR
- Include both unit tests and integration tests where appropriate
- Test on multiple platforms if possible

### Documentation

- Update documentation for any new features or changes
- Follow the existing documentation style
- Include examples where appropriate
- Update README.md if necessary

## Project Structure

```
.
├── cmd/              # Command-line interface
├── config/           # Configuration management
├── manager/          # Core functionality
├── tests/            # Test files
├── main.go          # Entry point
├── go.mod           # Go module file
└── go.sum           # Go module checksum
```

## Release Process

1. Update version numbers in relevant files
2. Update CHANGELOG.md
3. Create a new release tag
4. Build and test the release
5. Create a GitHub release

## Getting Help

- Check the [documentation](https://github.com/Snupai/cli-config-manager/wiki)
- Open an issue for questions
- Join our community discussions

## License

By contributing to dotman, you agree that your contributions will be licensed under the project's MIT License. 
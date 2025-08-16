# Contributing to Neon

Thank you for your interest in contributing to Neon! We welcome contributions from the community and are pleased to have you join us.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Issue Reporting](#issue-reporting)
- [Security Vulnerabilities](#security-vulnerabilities)

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

### Our Standards

- **Be respectful**: Treat all community members with respect and kindness
- **Be inclusive**: Welcome newcomers and encourage diverse perspectives
- **Be collaborative**: Work together and help each other learn and grow
- **Be constructive**: Provide helpful feedback and suggestions

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Git
- A GitHub account

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/neon.git
   cd neon
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/sri-shubham/neon.git
   ```

## Development Setup

1. **Install Dependencies**:
   ```bash
   go mod download
   ```

2. **Run Tests**:
   ```bash
   go test ./...
   ```

3. **Run Integration Tests**:
   ```bash
   go test -tags=integration ./...
   ```

4. **Check Code Coverage**:
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

## How to Contribute

### Types of Contributions

We welcome several types of contributions:

- **Bug Reports**: Report bugs using GitHub issues
- **Feature Requests**: Suggest new features or improvements
- **Code Contributions**: Submit bug fixes, new features, or improvements
- **Documentation**: Improve existing documentation or add new docs
- **Examples**: Add usage examples or tutorials

### Before You Start

1. **Check existing issues**: Look for existing issues or pull requests
2. **Open an issue**: For significant changes, open an issue first to discuss
3. **Keep it focused**: Make small, focused changes rather than large, sweeping ones

## Pull Request Process

### Before Submitting

1. **Update your fork**:
   ```bash
   git fetch upstream
   git checkout master
   git merge upstream/master
   ```

2. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes** following our [coding standards](#coding-standards)

4. **Test your changes**:
   ```bash
   go test ./...
   go test -race ./...
   ```

5. **Update documentation** if necessary

### Submitting the Pull Request

1. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

2. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create a Pull Request** on GitHub with:
   - Clear title and description
   - Reference to related issues
   - Screenshots or examples if applicable

### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(middleware): add authentication middleware support
fix(router): resolve path parameter parsing issue
docs(readme): update installation instructions
test(integration): add comprehensive endpoint tests
```

## Coding Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` to format your code
- Use `golint` and `go vet` to check for issues
- Write clear, self-documenting code with meaningful names

### Code Quality Tools

Run these tools before submitting:

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run golangci-lint (if available)
golangci-lint run
```

### Code Organization

- Keep functions small and focused
- Use meaningful package and variable names
- Add comments for exported functions and types
- Group related functionality together

## Testing Guidelines

### Test Requirements

- **Unit Tests**: All new code must have unit tests
- **Integration Tests**: Add integration tests for new features
- **Coverage**: Maintain 100% test coverage where possible
- **Performance**: Include benchmarks for performance-critical code

### Test Structure

```go
func TestFeatureName(t *testing.T) {
    // Arrange
    // ... setup test data
    
    // Act
    // ... call the function being tested
    
    // Assert
    // ... verify the results
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestSpecificFunction

# Run benchmarks
go test -bench=.
```

## Documentation

### Documentation Standards

- **README**: Keep the main README.md updated
- **Code Comments**: Document all exported functions and types
- **Examples**: Include practical examples in documentation
- **API Docs**: Use godoc format for API documentation

### Updating Documentation

When making changes:

1. Update relevant documentation
2. Add examples for new features
3. Update the README.md if necessary
4. Ensure all code is properly commented

## Issue Reporting

### Bug Reports

When reporting bugs, please include:

- **Go version**: `go version`
- **Neon version**: Version you're using
- **Operating System**: OS and version
- **Description**: Clear description of the issue
- **Steps to Reproduce**: Minimal steps to reproduce the bug
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Code Sample**: Minimal code that reproduces the issue

### Feature Requests

For feature requests, please include:

- **Use Case**: Describe the problem you're trying to solve
- **Proposed Solution**: Your ideas for implementation
- **Alternatives**: Other solutions you've considered
- **Additional Context**: Any other relevant information

## Security Vulnerabilities

If you discover a security vulnerability, please:

1. **Do NOT** open a public GitHub issue
2. Email the maintainers directly
3. Provide detailed information about the vulnerability
4. Allow time for the issue to be addressed before public disclosure

## Getting Help

If you need help:

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Documentation**: Check the README and godoc

## Recognition

Contributors will be recognized in:

- **CHANGELOG.md**: Major contributions noted in release notes
- **README.md**: Contributors section (coming soon)
- **GitHub**: Contributor recognition features

## Development Workflow

### Typical Workflow

1. Check issues or create new one
2. Fork and clone the repository
3. Create a feature branch
4. Make changes with tests
5. Run all tests and checks
6. Update documentation
7. Submit pull request
8. Address review feedback
9. Merge after approval

### Release Process

Releases follow semantic versioning:

- **Major**: Breaking changes (1.0.0 → 2.0.0)
- **Minor**: New features (1.0.0 → 1.1.0)
- **Patch**: Bug fixes (1.0.0 → 1.0.1)

## Questions?

Don't hesitate to ask questions! We're here to help and want to make contributing as smooth as possible.

Thank you for contributing to Neon! 🚀

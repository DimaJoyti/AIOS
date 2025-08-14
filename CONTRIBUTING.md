# Contributing to AIOS

Thank you for your interest in contributing to AIOS (AI-Powered Operating System)! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Node.js 18 or higher
- Docker and Docker Compose
- Git
- Arch Linux (recommended for development)

### Development Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/aios/aios.git
   cd aios
   ```

2. **Run the setup script:**
   ```bash
   ./scripts/setup-dev.sh
   ```

3. **Start the development environment:**
   ```bash
   make dev
   ```

## Contributing Guidelines

### Types of Contributions

We welcome several types of contributions:

- **Bug Reports**: Help us identify and fix issues
- **Feature Requests**: Suggest new features or improvements
- **Code Contributions**: Submit bug fixes, new features, or improvements
- **Documentation**: Improve or add documentation
- **Testing**: Add or improve tests

### Before You Start

1. **Check existing issues**: Look for existing issues or discussions related to your contribution
2. **Create an issue**: For significant changes, create an issue to discuss your proposal
3. **Fork the repository**: Create a personal fork of the project
4. **Create a branch**: Create a feature branch for your changes

### Branch Naming Convention

Use descriptive branch names that follow this pattern:
- `feature/description-of-feature`
- `bugfix/description-of-bug`
- `docs/description-of-documentation`
- `refactor/description-of-refactor`

## Pull Request Process

### Before Submitting

1. **Test your changes**: Ensure all tests pass
2. **Update documentation**: Update relevant documentation
3. **Follow coding standards**: Ensure your code follows our standards
4. **Commit message format**: Use conventional commit messages

### Commit Message Format

We use conventional commits for clear and consistent commit messages:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

Examples:
```
feat(ai): add natural language command processing
fix(security): resolve authentication bypass vulnerability
docs(api): update API documentation for new endpoints
```

### Pull Request Template

When creating a pull request, please include:

1. **Description**: Clear description of what the PR does
2. **Related Issues**: Link to related issues
3. **Type of Change**: Bug fix, new feature, breaking change, etc.
4. **Testing**: How you tested your changes
5. **Screenshots**: If applicable, add screenshots

## Coding Standards

### Go Code Standards

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` and `goimports` for formatting
- Run `golangci-lint` for linting
- Write comprehensive tests for new code
- Use meaningful variable and function names
- Add comments for exported functions and complex logic

### TypeScript/React Standards

- Use TypeScript for all new frontend code
- Follow React best practices and hooks patterns
- Use functional components over class components
- Implement proper error boundaries
- Use Tailwind CSS for styling
- Follow accessibility guidelines (WCAG 2.1)

### General Standards

- **Error Handling**: Always handle errors appropriately
- **Logging**: Use structured logging with appropriate levels
- **Security**: Follow security best practices
- **Performance**: Consider performance implications
- **Documentation**: Document public APIs and complex logic

## Testing

### Go Testing

- Write unit tests for all new functions
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for >90% test coverage
- Run tests with: `make test`

### Frontend Testing

- Write unit tests for components and utilities
- Use React Testing Library for component tests
- Test user interactions and accessibility
- Run tests with: `npm test`

### Integration Testing

- Write integration tests for API endpoints
- Test complete user workflows
- Use Docker for consistent test environments
- Run integration tests with: `make test-integration`

## Documentation

### Code Documentation

- Document all exported functions and types
- Use GoDoc format for Go code
- Use JSDoc format for TypeScript code
- Include examples in documentation

### User Documentation

- Update README.md for user-facing changes
- Update API documentation for API changes
- Add or update tutorials for new features
- Keep documentation up to date with code changes

### Architecture Documentation

- Update ARCHITECTURE.md for architectural changes
- Document design decisions and trade-offs
- Include diagrams for complex systems
- Maintain ADRs (Architecture Decision Records)

## Development Workflow

### Local Development

1. **Start services**: `make dev`
2. **Run tests**: `make test`
3. **Check linting**: `make lint`
4. **Format code**: `make format`

### Debugging

- Use Delve for Go debugging
- Use browser dev tools for frontend debugging
- Check logs in development environment
- Use distributed tracing for complex issues

### Performance Testing

- Run benchmarks: `make benchmark`
- Profile CPU and memory usage
- Test under load conditions
- Monitor resource usage

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):
- MAJOR: Incompatible API changes
- MINOR: Backward-compatible functionality additions
- PATCH: Backward-compatible bug fixes

### Release Checklist

1. Update version numbers
2. Update CHANGELOG.md
3. Run full test suite
4. Update documentation
5. Create release notes
6. Tag the release
7. Deploy to staging
8. Deploy to production

## Getting Help

### Communication Channels

- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Discord**: For real-time chat and community support
- **Email**: For security issues and private matters

### Resources

- [Project Documentation](docs/)
- [API Documentation](docs/api/)
- [Architecture Guide](ARCHITECTURE.md)
- [Deployment Guide](docs/deployment.md)

## Recognition

Contributors will be recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project documentation
- Annual contributor highlights

Thank you for contributing to AIOS! Your contributions help make AI-powered computing accessible to everyone.

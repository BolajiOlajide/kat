---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql,contributing
title: Contributing
description: |
    Learn how to contribute to the Kat project.
comments: false
permalink: /contributing/
page_nav:
    prev:
        content: Update
        url: '/update'
---

# Contributing to Kat

Contributions to Kat are welcome and appreciated! This guide will help you get started with contributing to the project.

## Getting Started

### Prerequisites

- Go 1.15 or higher
- Git
- PostgreSQL (for testing)

### Setting Up the Development Environment

```bash
# Clone the repository
git clone https://github.com/BolajiOlajide/kat.git
cd kat

# Install dependencies
go mod download
```

## Development Workflow

1. **Fork the repository** on GitHub
2. **Create your feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Commit your changes**
   ```bash
   git commit -m 'Add some amazing feature'
   ```
4. **Push to the branch**
   ```bash
   git push origin feature/amazing-feature
   ```
5. **Open a Pull Request**

## Building Locally

To build Kat locally:

```bash
make build
```

This will create a binary in the current directory.

## Running Tests

To run the tests:

```bash
make test
```

Make sure you have a PostgreSQL instance available for integration tests. You can configure the test database connection using environment variables.

## Code Style

- Follow standard Go code formatting guidelines
- Use `goimports` to organize imports (available in the `.bin` directory)
- Run `go fmt ./...` before committing

## Pull Request Process

1. Update the README.md or documentation with details of changes if appropriate
2. Update the version number in version files following Semantic Versioning
3. Your PR will be reviewed by maintainers, who may request changes
4. Once approved, your PR will be merged

## Reporting Bugs

When reporting bugs, please include:

- A clear description of the issue
- Steps to reproduce
- Expected behavior
- Actual behavior
- Kat version and environment details

## Feature Requests

Feature requests are welcome! Please provide:

- A clear description of the feature
- The problem it solves
- Any design ideas or considerations

## Communication

- Use GitHub Issues for bug reports and feature requests
- For questions, you can open a Discussion on GitHub

## Releasing

Kat uses [GoReleaser](https://goreleaser.com/) for building and publishing releases. Maintainers will handle the release process.

## License

By contributing to Kat, you agree that your contributions will be licensed under the project's [Apache License 2.0](https://github.com/BolajiOlajide/kat/blob/main/LICENSE).
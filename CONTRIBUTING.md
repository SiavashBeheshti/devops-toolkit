# Contributing to DevOps Toolkit

First off, thank you for considering contributing to DevOps Toolkit! It's people like you that make DevOps Toolkit such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### 🐛 Reporting Bugs

Before creating bug reports, please check the existing issues as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

**Bug Report Template:**

```markdown
**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Run command '...'
2. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots/Output**
If applicable, add screenshots or command output to help explain your problem.

**Environment:**
 - OS: [e.g., Ubuntu 22.04, macOS 14]
 - Go Version: [e.g., 1.21]
 - DevOps Toolkit Version: [e.g., 0.1.0]
 - Kubernetes Version: [if applicable]
 - Docker Version: [if applicable]

**Additional context**
Add any other context about the problem here.
```

### 💡 Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

- **Use a clear and descriptive title**
- **Provide a detailed description** of the suggested enhancement
- **Explain why this enhancement would be useful**
- **List any alternatives you've considered**

### 🔧 Pull Requests

1. **Fork the repo** and create your branch from `main`
2. **Follow the coding style** used throughout the project
3. **Add tests** for any new functionality
4. **Ensure the test suite passes** (`make test`)
5. **Run the linter** (`make lint`)
6. **Update documentation** as needed
7. **Write a clear commit message**

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Docker (for testing Docker commands)
- kubectl configured with cluster access (for testing K8s commands)
- golangci-lint (for linting)

### Getting Started

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/devops-toolkit.git
cd devops-toolkit

# Add upstream remote
git remote add upstream https://github.com/beheshti/devops-toolkit.git

# Install dependencies
go mod download

# Build the project
make build

# Run tests
make test

# Run linter
make lint
```

### Project Structure

```
devops-toolkit/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── k8s/               # Kubernetes commands
│   ├── docker/            # Docker commands
│   ├── gitlab/            # GitLab commands
│   └── compliance/        # Compliance commands
├── pkg/                    # Reusable packages
│   ├── output/            # Terminal output utilities
│   ├── k8s/               # Kubernetes client
│   ├── docker/            # Docker client
│   ├── gitlabclient/      # GitLab client
│   └── compliance/        # Compliance engine
├── main.go
├── go.mod
└── Makefile
```

## Coding Guidelines

### Go Style

- Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` for formatting (run `make fmt`)
- Write clear, self-documenting code
- Add comments for exported functions and types

### Commit Messages

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that don't affect code meaning
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding missing tests
- `chore`: Changes to build process or auxiliary tools

**Examples:**
```
feat(k8s): add pod log streaming support
fix(docker): handle container names with special characters
docs: update installation instructions
refactor(output): simplify table rendering logic
```

### Adding New Commands

1. Create a new file in the appropriate `cmd/` subdirectory
2. Use the existing command structure as a template
3. Add the command to the parent command in the main file
4. Write tests for the new command
5. Update documentation

**Command Template:**

```go
package mycommand

import (
    "github.com/beheshti/devops-toolkit/pkg/output"
    "github.com/spf13/cobra"
)

func newMyCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mycommand",
        Short: "Brief description",
        Long:  `Detailed description of the command.`,
        RunE:  runMyCommand,
    }

    // Add flags
    cmd.Flags().StringP("flag", "f", "", "Flag description")

    return cmd
}

func runMyCommand(cmd *cobra.Command, args []string) error {
    output.StartSpinner("Doing something...")
    
    // Implementation
    
    output.SpinnerSuccess("Done!")
    return nil
}
```

### Output Guidelines

Use the `pkg/output` package for consistent, beautiful output:

```go
// Spinners for async operations
output.StartSpinner("Loading...")
output.SpinnerSuccess("Done!")
output.SpinnerError("Failed!")

// Status messages
output.Success("Operation completed")
output.Warning("Something might be wrong")
output.Error("Operation failed")
output.Info("Here's some info")

// Tables
table := output.NewTable(output.TableConfig{
    Title:      "My Table",
    Headers:    []string{"Col1", "Col2"},
    ShowBorder: true,
})
table.AddRow([]string{"value1", "value2"})
table.Render()

// Colored rows based on status
row, colors := output.StatusRow("Component", "Healthy", "Details")
table.AddColoredRow(row, colors)
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific tests
go test -v ./pkg/k8s/...
```

### Writing Tests

- Place tests in `_test.go` files next to the code
- Use table-driven tests for multiple scenarios
- Mock external dependencies

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "result",
            wantErr:  false,
        },
        // Add more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.expected {
                t.Errorf("MyFunction() = %v, want %v", got, tt.expected)
            }
        })
    }
}
```

## Review Process

1. All submissions require review
2. Maintainers will review your PR within a few days
3. Address any feedback or requested changes
4. Once approved, your PR will be merged

## Recognition

Contributors will be recognized in:
- The project's README
- Release notes for features/fixes they contributed
- The GitHub contributors page

## Questions?

Feel free to:
- Open an issue for questions
- Start a discussion in GitHub Discussions
- Reach out to maintainers

Thank you for contributing! 🎉


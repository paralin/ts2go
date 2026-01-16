# ts2go

**ts2go** is an experimental TypeScript-to-Go transpiler that translates TypeScript code to Go at the AST level.

## Overview

This project creates a transpiler that converts TypeScript source code into equivalent Go code. It uses the TypeScript compiler API for parsing and implements a custom code generator to produce idiomatic Go code.

## Features

- **AST-Level Transpilation**: Parses TypeScript using the official TypeScript compiler API and generates Go code from the abstract syntax tree
- **Compliance Test Suite**: Comprehensive test suite using Go's testing framework to ensure correctness
- **Basic TypeScript Support**: Handles variables, functions, basic expressions, and more

## Installation

### Prerequisites

- Go 1.23 or later
- Node.js 18 or later
- npm or yarn

### Building from Source

```bash
# Clone the repository
git clone https://github.com/paralin/ts2go.git
cd ts2go

# Install Node.js dependencies (TypeScript parser)
npm install

# Build the CLI tool
go build -o ts2go ./cmd/ts2go
```

## Usage

### Command Line

Compile a TypeScript file to Go:

```bash
./ts2go -input example.ts -output ./output
```

Options:
- `-input`: Path to the input TypeScript file (required)
- `-output`: Output directory for generated Go files (default: `./output`)
- `-verbose`: Enable verbose logging

### Example

**Input TypeScript** (`example.ts`):
```typescript
const x = 10;
const y = 20;
const result = x + y;
console.log("Result:", result);
```

**Generated Go** (`example.go`):
```go
package main

import "fmt"

func main() {
	x := 10
	y := 20
	result := x + y
	fmt.Println("Result:", result)
}
```

## Testing

The project includes a compliance test suite that verifies TypeScript code transpiles correctly and produces the same output as the original TypeScript.

### Running Tests

```bash
# Run all compliance tests
go test -v ./tests

# Run a specific test
go test -v ./tests -run TestCompliance/basic_arithmetic
```

### Test Structure

Each test is a directory under `tests/tests/` containing:
- `*.ts` - TypeScript source files
- `expected.log` - Expected output (auto-generated if missing)
- `actual.log` - Actual output when tests fail (for debugging)

Optional marker files:
- `skip-test` - Skip this test
- `expect-fail` - Expect compilation to fail

## Project Structure

```
ts2go/
├── cmd/
│   └── ts2go/          # CLI tool
├── compiler/           # Transpiler implementation
│   ├── compiler.go     # Main compiler logic
│   ├── ast.go          # AST type definitions
│   └── codegen.go      # Go code generation
├── tests/              # Test infrastructure
│   ├── tests/          # Compliance test cases
│   └── tests.go        # Test runner
├── parser.js           # TypeScript parser wrapper
├── package.json        # Node.js dependencies
└── go.mod             # Go module definition
```

## Architecture

1. **TypeScript Parsing**: Uses the TypeScript compiler API (via `parser.js`) to parse TypeScript source into an AST
2. **AST Transformation**: Processes the TypeScript AST and maps nodes to Go equivalents
3. **Code Generation**: Generates idiomatic Go code with proper formatting and syntax

## Current Limitations

This is an experimental project with the following limitations:

- Limited TypeScript feature support (basic variables, functions, expressions)
- No type system mapping yet
- No generics support
- No class/interface transpilation
- Basic expression handling only

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## Acknowledgments

This project is inspired by [aperturerobotics/goscript](https://github.com/aperturerobotics/goscript), which transpiles Go to TypeScript.

## License

MIT License - see [LICENSE](LICENSE) file for details.

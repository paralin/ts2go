# Agent Rules for ts2go

This document contains guidelines and rules for AI agents working on the ts2go project.

## IMPORTANT

Fix anything that you come across in the project while working that violates any of these guidelines as you encounter it.

**CRITICAL: When asked to update AGENTS.md:**

- ALWAYS read the ENTIRE file first before making edits
- Check for duplicate information across sections
- Condense and consolidate duplicates into a single authoritative section
- Ensure guidelines are clear and non-contradictory

Remember to always delete dead code when changing things - for example if you changed something and a function is no longer used anywhere, delete that function.

### General Rules

- Try to keep things in one function unless composable or reusable
- PREFER one exported struct per `.go` file (file named after the struct, e.g., `compiler.go` for `Compiler`)
  - Multiple unexported (internal) structs in the same file are acceptable
  - Constants and type aliases can be co-located with the struct that uses them
- DO NOT use `else` statements unless necessary
- DO NOT make git commits
- AVOID `else` statements or "fallback" cases
- PREFER single word variable names where possible
- DO NOT maintain backwards compatibility - this is an experimental project
- Remove any "for backwards compatibility" comments and fallback logic
- NEVER hardcode things: examples include function names, builtins, etc.
- Actively try to improve the codebase to conform to the above when the opportunity arises
- Go standard library sources are located at "go env GOROOT" (shell command)
- Leverage adding more tests, for example in `tests/tests/`, instead of debug logging, for diagnosing issues or investigating hypotheses
- AVOID type arguments unless necessary (prefer type inference)
- When making Git commits use the existing commit message pattern and Linux-kernel style commit message bodies
- When you would normally add a new compliance test check if a very-similar compliance test already exists and if so extend that one instead

## Project Overview

ts2go is an experimental TypeScript to Go transpiler that converts TypeScript code into equivalent Go code at the AST level. It uses Microsoft's typescript-go library for parsing TypeScript and implements a custom code generator to produce idiomatic Go code.

**This is an experimental project** - we do not maintain backwards compatibility and prioritize simplicity and correctness over legacy support. You may sometimes encounter a problem that requires a complete re-design or re-think or re-architecting of an aspect of ts2go, which is perfectly okay, in this case write a design to `tests/WIP.md` and think it through extensively before performing your refactor. It's perfectly OK to delete large swaths of code as needed. Focus on correctness.

If you want to overwrite WIP.md you must `rm` it first.

**Output Style**: Generated Go code should follow standard Go formatting conventions and use `gofmt` style.

**Philosophy**: Follow Rick Rubin's concept of being both an engineer and a reducer (not always a producer) by focusing on the shortest, most straightforward solution that is correct.

## Compliance Testing Workflow

When working on compliance tests:

1. **Test Location**: Compliance tests are located at `./tests/tests/{testname}/main.ts` with TypeScript source code and corresponding expected output in `expected.log`.

2. **Running Tests**:

   **For a specific test:**
   ```bash
   go test -timeout 60s -run ^TestCompliance/basic_arithmetic$ ./tests
   ```

   **For the full suite:**
   ```bash
   # Run once, capture to file, check result
   mkdir -p .tmp && go test -timeout 10m ./tests 2>&1 > .tmp/test_output.txt; echo "Exit code: $?"

   # If exit code is non-zero, find all failing tests:
   grep -E "^--- FAIL:" .tmp/test_output.txt

   # Then run specific failing tests with -v for details:
   go test -v -timeout 60s -run ^TestCompliance/failing_test_name$ ./tests
   ```

   **IMPORTANT:** Do NOT pipe test output directly to grep/tail during the test run. The test framework may produce verbose output that looks like errors but isn't. Always check the exit code first, then analyze the output file if needed. The `.tmp/` directory is gitignored.

3. **Analysis Process**:
   - Run the compliance test to check if it passes
   - If not, review the output to see why
   - Deeply consider the generated Go code from the source TypeScript code
   - Think about what the correct Go output would look like with as minimal of a change as possible
   - **If the test is too complex** (many cascading errors, large dependencies, or unclear root cause):
     - Create a new, simpler compliance test that isolates a specific subset of the problem
     - Name it descriptively (e.g., `variable_declaration` for variable declaration issues)
     - Focus on reproducing just one aspect of the failure in minimal code
     - Fix the isolated test first, then return to the original complex test

4. **Implementation Workflow**:
   - Review the code under `compiler/*.go` to determine what needs to be changed
   - Write your analysis and info about the task at hand to `tests/WIP.md` (overwrite any existing contents)
   - Apply the planned changes to the `compiler/` code
   - Run the compliance test again
   - Repeat: update compiler code and/or `tests/WIP.md` until the compliance test passes successfully
   - If you make two or more edits and the test still does not pass, ask the user how to proceed providing several options
   - After fixing a specific test, re-run the full compliance test to verify everything works properly

Once the issue is fixed and the compliance test passes you may delete WIP.md without updating it with a final summary.

## Design Patterns & Code Style

### Core Principles

1. **Function Naming Convention**: When writing functions that convert TypeScript AST to Go, name them descriptively:
   - For variable declarations, use `visitVariableDeclaration` or `generateVariableDeclaration`
   - Try to make a clear match between AST type and function name
   - Avoid hiding logic in unexported functions unless necessary

2. **Function Organization**:
   - Avoid splitting functions unless:
     - The logic is reused elsewhere
     - The function becomes excessively long and complex
     - Doing so adheres to existing patterns in the codebase

3. **Implementation Completeness**:
   - Avoid leaving undecided implementation details in the code
   - Make a decision and add a comment explaining the choice if necessary

4. **Struct Field Policy**:
   - You **may** add new fields to `Compiler` if necessary
   - You **may** add new fields to `codeGenerator` for tracking state during generation

## Linting and Code Quality

When working with code quality:

1. **Running Tests**: Use `go test ./...` to run all tests
2. **Formatting**: Use `go fmt ./...` or `gofmt` to format code
3. **Fixing Errors**: Address compiler errors in the affected code files
4. **Iterating**: Repeat the testing process until no errors remain
5. **Ignoring Warnings**: You can ignore linter errors with inline comments when the warning is unnecessarily strict:
   ```go
   defer f.Close() //nolint:errcheck
   ```
   This is appropriate for cases like deferring Close without checking the error return value, which can often be safely ignored.

Make sure that `go test ./...` passes before suggesting a task is complete.

## Comments

When adding comments to functions, structs, or files:

- Use the format: `// FunctionName does something specific.`
- Start with the function/struct/type name followed by a verb
- End with a period
- Keep it concise and descriptive

Example:

```go
// visitVariableDeclaration converts a TypeScript variable declaration to Go.
func (g *codeGenerator) visitVariableDeclaration(decl *ast.VariableDeclaration) error {
  ...
}
```

### Go Type Assertions

When using type assertion syntax in Go, add a comment line just before:

```go
// _ is a type assertion
var _ SomeInterface = (*SomeStruct)(nil)
```

This verifies at compile-time that `SomeStruct` implements `SomeInterface`.

## TypeScript-Go Library Usage

Since microsoft/typescript-go uses internal packages that cannot be imported directly:

1. **DO NOT** import from `github.com/microsoft/typescript-go/internal/*` packages
2. **Find and use** public APIs when available
3. **If no public API exists**, consider:
   - Wrapping the functionality we need
   - Using a different approach
   - Vendoring specific code if absolutely necessary
4. **Document workarounds** clearly in comments

## Key Differences from GoScript

ts2go is the inverse of GoScript:
- **Input**: TypeScript source code
- **Output**: Go source code
- **Parser**: Uses microsoft/typescript-go for TypeScript parsing
- **Generator**: Produces idiomatic Go code from TypeScript AST
- **Testing**: Validates that generated Go produces same output as original TypeScript

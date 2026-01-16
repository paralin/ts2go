package tests

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paralin/ts2go/compiler"
	"github.com/sirupsen/logrus"
)

// TestCompliance runs all compliance tests in the tests/tests directory.
func TestCompliance(t *testing.T) {
	testsDir := filepath.Join(".", "tests")
	
	entries, err := os.ReadDir(testsDir)
	if err != nil {
		t.Fatalf("Failed to read tests directory: %v", err)
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		testName := entry.Name()
		testDir := filepath.Join(testsDir, testName)
		
		// Check if test should be skipped
		if _, err := os.Stat(filepath.Join(testDir, "skip-test")); err == nil {
			t.Logf("Skipping test: %s (skip-test marker found)", testName)
			continue
		}
		
		t.Run(testName, func(t *testing.T) {
			runComplianceTest(t, testDir)
		})
	}
}

// runComplianceTest runs a single compliance test.
func runComplianceTest(t *testing.T, testDir string) {
	t.Helper()
	
	// Find TypeScript source files in the test directory
	tsFiles, err := filepath.Glob(filepath.Join(testDir, "*.ts"))
	if err != nil {
		t.Fatalf("Failed to find TypeScript files: %v", err)
	}
	
	if len(tsFiles) == 0 {
		t.Fatal("No TypeScript files found in test directory")
	}
	
	// Create output directory
	outputDir := filepath.Join(testDir, "output")
	if err := os.RemoveAll(outputDir); err != nil {
		t.Fatalf("Failed to remove old output directory: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	
	// Create compiler
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	
	conf := &compiler.Config{
		OutputPath: outputDir,
	}
	
	comp, err := compiler.NewCompiler(conf, logger)
	if err != nil {
		t.Fatalf("Failed to create compiler: %v", err)
	}
	
	// Compile each TypeScript file
	ctx := context.Background()
	for _, tsFile := range tsFiles {
		if err := comp.CompileFile(ctx, tsFile); err != nil {
			// Check if test expects compilation to fail
			if _, err := os.Stat(filepath.Join(testDir, "expect-fail")); err == nil {
				t.Logf("Compilation failed as expected: %v", err)
				return
			}
			t.Fatalf("Compilation failed: %v", err)
		}
	}
	
	// If test expects failure, but compilation succeeded, fail the test
	if _, err := os.Stat(filepath.Join(testDir, "expect-fail")); err == nil {
		t.Fatal("Expected compilation to fail, but it succeeded")
	}
	
	// Run the generated Go code
	goFiles, err := filepath.Glob(filepath.Join(outputDir, "*.go"))
	if err != nil {
		t.Fatalf("Failed to find generated Go files: %v", err)
	}
	
	if len(goFiles) == 0 {
		t.Fatal("No Go files were generated")
	}
	
	// Run the Go code and capture output
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("go", "run", goFiles[0])
	cmd.Dir = outputDir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run generated Go code: %v\nStderr: %s", err, stderr.String())
	}
	
	actualOutput := stdout.String()
	
	// Check if expected.log exists
	expectedLogPath := filepath.Join(testDir, "expected.log")
	expectedOutput := ""
	
	if _, err := os.Stat(expectedLogPath); err == nil {
		// Read expected output
		data, err := os.ReadFile(expectedLogPath)
		if err != nil {
			t.Fatalf("Failed to read expected.log: %v", err)
		}
		expectedOutput = string(data)
	} else {
		// Generate expected output by running the TypeScript
		t.Logf("Generating expected output from TypeScript...")
		tsOutput, err := runTypeScript(tsFiles[0])
		if err != nil {
			t.Fatalf("Failed to generate expected output: %v", err)
		}
		expectedOutput = tsOutput
		
		// Write expected.log
		if err := os.WriteFile(expectedLogPath, []byte(expectedOutput), 0644); err != nil {
			t.Fatalf("Failed to write expected.log: %v", err)
		}
	}
	
	// Compare outputs
	if normalizeOutput(actualOutput) != normalizeOutput(expectedOutput) {
		// Write actual.log for debugging
		actualLogPath := filepath.Join(testDir, "actual.log")
		os.WriteFile(actualLogPath, []byte(actualOutput), 0644)
		
		t.Fatalf("Output mismatch!\nExpected:\n%s\n\nActual:\n%s", expectedOutput, actualOutput)
	}
}

// runTypeScript runs a TypeScript file using ts-node or similar and returns the output.
func runTypeScript(tsFile string) (string, error) {
	var stdout, stderr bytes.Buffer
	
	// Try running with ts-node
	cmd := exec.Command("npx", "ts-node", tsFile)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run TypeScript: %v\nStderr: %s", err, stderr.String())
	}
	
	return stdout.String(), nil
}

// normalizeOutput normalizes output for comparison by trimming whitespace.
func normalizeOutput(s string) string {
	lines := strings.Split(s, "\n")
	var normalized []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return strings.Join(normalized, "\n")
}

package compiler

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Config holds the configuration for the TypeScript to Go compiler.
type Config struct {
	// OutputPath is the directory where the generated Go files will be written.
	OutputPath string
	// ParserScript is the path to the TypeScript parser script.
	// Defaults to "parser.js" in the project root.
	ParserScript string
}

// Compiler is the main TypeScript to Go compiler.
type Compiler struct {
	conf   *Config
	logger logrus.FieldLogger
}

// NewCompiler creates a new TypeScript to Go compiler instance.
func NewCompiler(conf *Config, logger logrus.FieldLogger) (*Compiler, error) {
	if conf.OutputPath == "" {
		return nil, errors.New("output path is required")
	}

	if conf.ParserScript == "" {
		conf.ParserScript = "parser.js"
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &Compiler{
		conf:   conf,
		logger: logger,
	}, nil
}

// CompileFile compiles a single TypeScript file to Go.
func (c *Compiler) CompileFile(ctx context.Context, inputPath string) error {
	c.logger.Infof("compiling TypeScript file: %s", inputPath)

	// Parse the TypeScript file
	ast, err := c.parseTypeScript(ctx, inputPath)
	if err != nil {
		return errors.Wrap(err, "failed to parse TypeScript")
	}

	// Generate Go code from the AST
	goCode, err := c.generateGoCode(ast)
	if err != nil {
		return errors.Wrap(err, "failed to generate Go code")
	}

	// Write the output file
	outputPath := c.getOutputPath(inputPath)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return errors.Wrap(err, "failed to create output directory")
	}

	if err := os.WriteFile(outputPath, []byte(goCode), 0644); err != nil {
		return errors.Wrap(err, "failed to write output file")
	}

	c.logger.Infof("generated Go file: %s", outputPath)
	return nil
}

// parseTypeScript parses a TypeScript file using the Node.js parser script.
func (c *Compiler) parseTypeScript(ctx context.Context, inputPath string) (*ASTNode, error) {
	cmd := exec.CommandContext(ctx, "node", c.conf.ParserScript, inputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse TypeScript file: %s", string(output))
	}

	var ast ASTNode
	if err := json.Unmarshal(output, &ast); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal AST")
	}

	return &ast, nil
}

// getOutputPath determines the output file path for a given input file.
func (c *Compiler) getOutputPath(inputPath string) string {
	baseName := filepath.Base(inputPath)
	// Change .ts extension to .go
	if ext := filepath.Ext(baseName); ext == ".ts" {
		baseName = baseName[:len(baseName)-3] + ".go"
	} else {
		baseName = baseName + ".go"
	}
	return filepath.Join(c.conf.OutputPath, baseName)
}

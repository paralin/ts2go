package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/paralin/ts2go/compiler"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		inputFile  = flag.String("input", "", "Input TypeScript file to compile")
		outputDir  = flag.String("output", "./output", "Output directory for generated Go files")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)
	
	flag.Parse()
	
	// Set up logging
	logger := logrus.New()
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
	
	// Validate input
	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -input flag is required\n")
		flag.Usage()
		os.Exit(1)
	}
	
	// Create compiler
	conf := &compiler.Config{
		OutputPath: *outputDir,
	}
	
	comp, err := compiler.NewCompiler(conf, logger)
	if err != nil {
		logger.Fatalf("Failed to create compiler: %v", err)
	}
	
	// Compile the file
	ctx := context.Background()
	if err := comp.CompileFile(ctx, *inputFile); err != nil {
		logger.Fatalf("Compilation failed: %v", err)
	}
	
	logger.Infof("Compilation successful!")
}

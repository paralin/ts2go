package compiler

import (
	"fmt"
	"strings"
)

// generateGoCode generates Go code from a TypeScript AST.
func (c *Compiler) generateGoCode(ast *ASTNode) (string, error) {
	gen := &codeGenerator{
		compiler: c,
		builder:  &strings.Builder{},
		indent:   0,
	}
	
	if err := gen.visitNode(ast); err != nil {
		return "", err
	}
	
	return gen.builder.String(), nil
}

// codeGenerator handles the actual code generation process.
type codeGenerator struct {
	compiler *Compiler
	builder  *strings.Builder
	indent   int
}

// visitNode processes a single AST node and generates corresponding Go code.
func (g *codeGenerator) visitNode(node *ASTNode) error {
	switch node.Kind {
	case "SourceFile":
		return g.visitSourceFile(node)
	case "VariableStatement", "FirstStatement":
		return g.visitVariableStatement(node)
	case "FunctionDeclaration":
		return g.visitFunctionDeclaration(node)
	case "ExpressionStatement":
		return g.visitExpressionStatement(node)
	case "ReturnStatement":
		return g.visitReturnStatement(node)
	case "Block":
		return g.visitBlock(node)
	default:
		// For unsupported nodes, just traverse children
		for _, child := range node.Children {
			if err := g.visitNode(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

// visitSourceFile processes the source file root node.
func (g *codeGenerator) visitSourceFile(node *ASTNode) error {
	g.writeLine("package main")
	g.writeLine("")
	
	// Process all children except EndOfFileToken
	for _, child := range node.Children {
		if child.Kind != "EndOfFileToken" {
			if err := g.visitNode(&child); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// visitVariableStatement processes variable declarations.
func (g *codeGenerator) visitVariableStatement(node *ASTNode) error {
	// TypeScript: const x = 5; or let x: number = 5;
	// Go: x := 5 or var x int = 5
	
	// Find the VariableDeclarationList child
	for _, child := range node.Children {
		if child.Kind == "VariableDeclarationList" {
			for _, decl := range child.Children {
				if decl.Kind == "VariableDeclaration" {
					if err := g.visitVariableDeclaration(&decl); err != nil {
						return err
					}
				}
			}
		}
	}
	g.writeLine("")
	return nil
}

// visitVariableDeclaration processes a single variable declaration.
func (g *codeGenerator) visitVariableDeclaration(node *ASTNode) error {
	varName := node.Name
	
	// Find the initializer (the value being assigned)
	var initValue string
	var hasInit bool
	
	for _, child := range node.Children {
		if child.Kind != "Identifier" && child.Kind != "NumberKeyword" && 
		   child.Kind != "StringKeyword" && child.Kind != "BooleanKeyword" {
			// This is likely the initializer
			val, err := g.generateExpression(&child)
			if err != nil {
				return err
			}
			initValue = val
			hasInit = true
			break
		}
	}
	
	if hasInit {
		g.writeIndent()
		g.builder.WriteString(fmt.Sprintf("%s := %s\n", varName, initValue))
	} else {
		g.writeIndent()
		g.builder.WriteString(fmt.Sprintf("var %s\n", varName))
	}
	
	return nil
}

// visitFunctionDeclaration processes function declarations.
func (g *codeGenerator) visitFunctionDeclaration(node *ASTNode) error {
	funcName := node.Name
	
	// For now, generate a simple function without parameters or return types
	g.writeIndent()
	g.builder.WriteString(fmt.Sprintf("func %s() {\n", funcName))
	g.indent++
	
	// Find and process the function body
	for _, child := range node.Children {
		if child.Kind == "Block" {
			if err := g.visitBlock(&child); err != nil {
				return err
			}
		}
	}
	
	g.indent--
	g.writeIndent()
	g.builder.WriteString("}\n\n")
	
	return nil
}

// visitBlock processes a block statement.
func (g *codeGenerator) visitBlock(node *ASTNode) error {
	// Don't write the braces, they're handled by the parent
	for _, child := range node.Children {
		if child.Kind != "FirstPunctuation" && child.Kind != "CloseBraceToken" {
			if err := g.visitNode(&child); err != nil {
				return err
			}
		}
	}
	return nil
}

// visitExpressionStatement processes expression statements.
func (g *codeGenerator) visitExpressionStatement(node *ASTNode) error {
	for _, child := range node.Children {
		expr, err := g.generateExpression(&child)
		if err != nil {
			return err
		}
		if expr != "" {
			g.writeIndent()
			g.builder.WriteString(expr)
			g.builder.WriteString("\n")
		}
	}
	return nil
}

// visitReturnStatement processes return statements.
func (g *codeGenerator) visitReturnStatement(node *ASTNode) error {
	g.writeIndent()
	g.builder.WriteString("return")
	
	// Find the return value expression
	for _, child := range node.Children {
		if child.Kind != "ReturnKeyword" {
			expr, err := g.generateExpression(&child)
			if err != nil {
				return err
			}
			if expr != "" {
				g.builder.WriteString(" ")
				g.builder.WriteString(expr)
			}
		}
	}
	
	g.builder.WriteString("\n")
	return nil
}

// generateExpression generates code for an expression node.
func (g *codeGenerator) generateExpression(node *ASTNode) (string, error) {
	switch node.Kind {
	case "NumericLiteral", "FirstLiteralToken":
		return node.Text, nil
	case "StringLiteral":
		return node.Text, nil
	case "TrueKeyword":
		return "true", nil
	case "FalseKeyword":
		return "false", nil
	case "Identifier":
		return node.Name, nil
	case "BinaryExpression":
		return g.generateBinaryExpression(node)
	case "CallExpression":
		return g.generateCallExpression(node)
	case "PropertyAccessExpression":
		return g.generatePropertyAccessExpression(node)
	default:
		// Return the text as-is for unknown expressions
		return node.Text, nil
	}
}

// generateBinaryExpression generates code for binary expressions (e.g., a + b).
func (g *codeGenerator) generateBinaryExpression(node *ASTNode) (string, error) {
	if len(node.Children) < 3 {
		return node.Text, nil
	}
	
	left, err := g.generateExpression(&node.Children[0])
	if err != nil {
		return "", err
	}
	
	operator := node.Children[1].Text
	
	right, err := g.generateExpression(&node.Children[2])
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%s %s %s", left, operator, right), nil
}

// generateCallExpression generates code for function calls.
func (g *codeGenerator) generateCallExpression(node *ASTNode) (string, error) {
	if len(node.Children) == 0 {
		return node.Text, nil
	}
	
	// First child is the function name/expression
	funcName, err := g.generateExpression(&node.Children[0])
	if err != nil {
		return "", err
	}
	
	// Map console.log to fmt.Println
	if funcName == "console.log" {
		funcName = "fmt.Println"
	}
	
	// Collect arguments - they are direct children after the function name
	args := []string{}
	for i := 1; i < len(node.Children); i++ {
		child := &node.Children[i]
		// Skip punctuation tokens
		if child.Kind != "OpenParenToken" && child.Kind != "CloseParenToken" && 
		   child.Kind != "CommaToken" {
			argStr, err := g.generateExpression(child)
			if err != nil {
				return "", err
			}
			args = append(args, argStr)
		}
	}
	
	return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", ")), nil
}

// generatePropertyAccessExpression generates code for property access (e.g., obj.prop).
func (g *codeGenerator) generatePropertyAccessExpression(node *ASTNode) (string, error) {
	parts := []string{}
	for _, child := range node.Children {
		if child.Kind == "Identifier" {
			parts = append(parts, child.Name)
		}
	}
	return strings.Join(parts, "."), nil
}

// writeIndent writes the current indentation level.
func (g *codeGenerator) writeIndent() {
	for i := 0; i < g.indent; i++ {
		g.builder.WriteString("\t")
	}
}

// writeLine writes a line with proper indentation.
func (g *codeGenerator) writeLine(line string) {
	g.writeIndent()
	g.builder.WriteString(line)
	g.builder.WriteString("\n")
}

package compiler

import (
	"fmt"
	"strings"

	"github.com/paralin/typescript-go/ts2go"
)

// generateGoCode generates Go code from a TypeScript AST.
func (c *Compiler) generateGoCode(sourceFile *ast.SourceFile) (string, error) {
	gen := &codeGenerator{
		compiler: c,
		builder:  &strings.Builder{},
		indent:   0,
	}
	
	if err := gen.visitSourceFile(sourceFile); err != nil {
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

// visitSourceFile processes the source file root node.
func (g *codeGenerator) visitSourceFile(sourceFile *ast.SourceFile) error {
	g.writeLine("package main")
	g.writeLine("")
	
	// Process all statements in the source file
	if sourceFile.Statements != nil {
		for _, stmt := range sourceFile.Statements.Nodes {
			if err := g.visitNode(stmt); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// visitNode processes a single AST node and generates corresponding Go code.
func (g *codeGenerator) visitNode(node *ast.Node) error {
	if node == nil {
		return nil
	}
	
	switch node.Kind() {
	case ast.KindVariableStatement:
		return g.visitVariableStatement(node.AsVariableStatement())
	case ast.KindFunctionDeclaration:
		return g.visitFunctionDeclaration(node.AsFunctionDeclaration())
	case ast.KindExpressionStatement:
		return g.visitExpressionStatement(node.AsExpressionStatement())
	case ast.KindReturnStatement:
		return g.visitReturnStatement(node.AsReturnStatement())
	case ast.KindBlock:
		return g.visitBlock(node.AsBlock())
	default:
		// For unsupported nodes, silently skip
		g.compiler.logger.Debugf("Skipping unsupported node kind: %v", node.Kind())
	}
	return nil
}

// visitVariableStatement processes variable declarations.
func (g *codeGenerator) visitVariableStatement(stmt *ast.VariableStatement) error {
	if stmt == nil || stmt.DeclarationList == nil {
		return nil
	}
	
	declList := stmt.DeclarationList
	if declList.Declarations == nil {
		return nil
	}
	
	for _, decl := range declList.Declarations.Nodes {
		if decl == nil {
			continue
		}
		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil {
			continue
		}
		
		if err := g.visitVariableDeclaration(varDecl); err != nil {
			return err
		}
	}
	g.writeLine("")
	return nil
}

// visitVariableDeclaration processes a single variable declaration.
func (g *codeGenerator) visitVariableDeclaration(decl *ast.VariableDeclaration) error {
	if decl == nil || decl.Name == nil {
		return nil
	}
	
	// Get the variable name
	varName := g.getIdentifierText(decl.Name)
	
	// Check if there's an initializer
	if decl.Initializer != nil {
		initValue, err := g.generateExpression(decl.Initializer)
		if err != nil {
			return err
		}
		g.writeIndent()
		g.builder.WriteString(fmt.Sprintf("%s := %s\n", varName, initValue))
	} else {
		g.writeIndent()
		g.builder.WriteString(fmt.Sprintf("var %s\n", varName))
	}
	
	return nil
}

// visitFunctionDeclaration processes function declarations.
func (g *codeGenerator) visitFunctionDeclaration(decl *ast.FunctionDeclaration) error {
	if decl == nil || decl.Name == nil {
		return nil
	}
	
	funcName := g.getIdentifierText(decl.Name)
	
	// For now, generate a simple function without parameters or return types
	g.writeIndent()
	g.builder.WriteString(fmt.Sprintf("func %s() {\n", funcName))
	g.indent++
	
	// Process the function body
	if decl.Body != nil {
		if err := g.visitBlock(decl.Body); err != nil {
			return err
		}
	}
	
	g.indent--
	g.writeIndent()
	g.builder.WriteString("}\n\n")
	
	return nil
}

// visitBlock processes a block statement.
func (g *codeGenerator) visitBlock(block *ast.Block) error {
	if block == nil || block.Statements == nil {
		return nil
	}
	
	for _, stmt := range block.Statements.Nodes {
		if err := g.visitNode(stmt); err != nil {
			return err
		}
	}
	return nil
}

// visitExpressionStatement processes expression statements.
func (g *codeGenerator) visitExpressionStatement(stmt *ast.ExpressionStatement) error {
	if stmt == nil || stmt.Expression == nil {
		return nil
	}
	
	expr, err := g.generateExpression(stmt.Expression)
	if err != nil {
		return err
	}
	if expr != "" {
		g.writeIndent()
		g.builder.WriteString(expr)
		g.builder.WriteString("\n")
	}
	return nil
}

// visitReturnStatement processes return statements.
func (g *codeGenerator) visitReturnStatement(stmt *ast.ReturnStatement) error {
	if stmt == nil {
		return nil
	}
	
	g.writeIndent()
	g.builder.WriteString("return")
	
	// Check for return value
	if stmt.Expression != nil {
		expr, err := g.generateExpression(stmt.Expression)
		if err != nil {
			return err
		}
		if expr != "" {
			g.builder.WriteString(" ")
			g.builder.WriteString(expr)
		}
	}
	
	g.builder.WriteString("\n")
	return nil
}

// generateExpression generates code for an expression node.
func (g *codeGenerator) generateExpression(node *ast.Node) (string, error) {
	if node == nil {
		return "", nil
	}
	
	switch node.Kind() {
	case ast.KindNumericLiteral:
		lit := node.AsNumericLiteral()
		if lit != nil {
			return lit.Text, nil
		}
	case ast.KindStringLiteral:
		lit := node.AsStringLiteral()
		if lit != nil {
			return lit.Text, nil
		}
	case ast.KindTrueKeyword, ast.KindFalseKeyword:
		return node.Kind().String(), nil
	case ast.KindIdentifier:
		return g.getIdentifierText(node), nil
	case ast.KindBinaryExpression:
		return g.generateBinaryExpression(node.AsBinaryExpression())
	case ast.KindCallExpression:
		return g.generateCallExpression(node.AsCallExpression())
	case ast.KindPropertyAccessExpression:
		return g.generatePropertyAccessExpression(node.AsPropertyAccessExpression())
	}
	
	// Return empty for unknown expressions
	return "", nil
}

// generateBinaryExpression generates code for binary expressions (e.g., a + b).
func (g *codeGenerator) generateBinaryExpression(expr *ast.BinaryExpression) (string, error) {
	if expr == nil {
		return "", nil
	}
	
	left, err := g.generateExpression(expr.Left)
	if err != nil {
		return "", err
	}
	
	// Get operator token text
	operator := expr.OperatorToken.Kind().String()
	// Convert token name to operator symbol
	operator = g.tokenToOperator(expr.OperatorToken.Kind())
	
	right, err := g.generateExpression(expr.Right)
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("%s %s %s", left, operator, right), nil
}

// generateCallExpression generates code for function calls.
func (g *codeGenerator) generateCallExpression(expr *ast.CallExpression) (string, error) {
	if expr == nil || expr.Expression == nil {
		return "", nil
	}
	
	// Get function name/expression
	funcName, err := g.generateExpression(expr.Expression)
	if err != nil {
		return "", err
	}
	
	// Map console.log to fmt.Println
	if funcName == "console.log" {
		funcName = "fmt.Println"
	}
	
	// Collect arguments
	args := []string{}
	if expr.Arguments != nil {
		for _, arg := range expr.Arguments.Nodes {
			argStr, err := g.generateExpression(arg)
			if err != nil {
				return "", err
			}
			args = append(args, argStr)
		}
	}
	
	return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", ")), nil
}

// generatePropertyAccessExpression generates code for property access (e.g., obj.prop).
func (g *codeGenerator) generatePropertyAccessExpression(expr *ast.PropertyAccessExpression) (string, error) {
	if expr == nil {
		return "", nil
	}
	
	obj, err := g.generateExpression(expr.Expression)
	if err != nil {
		return "", err
	}
	
	prop := g.getIdentifierText(expr.Name)
	
	return fmt.Sprintf("%s.%s", obj, prop), nil
}

// getIdentifierText extracts text from an identifier node.
func (g *codeGenerator) getIdentifierText(node *ast.Node) string {
	if node == nil {
		return ""
	}
	
	if node.Kind() == ast.KindIdentifier {
		if ident := node.AsIdentifier(); ident != nil {
			return ident.EscapedText
		}
	}
	
	return ""
}

// tokenToOperator converts a token kind to its operator symbol.
func (g *codeGenerator) tokenToOperator(kind ast.Kind) string {
	switch kind {
	case ast.KindPlusToken:
		return "+"
	case ast.KindMinusToken:
		return "-"
	case ast.KindAsteriskToken:
		return "*"
	case ast.KindSlashToken:
		return "/"
	case ast.KindPercentToken:
		return "%"
	case ast.KindEqualsEqualsToken:
		return "=="
	case ast.KindExclamationEqualsToken:
		return "!="
	case ast.KindLessThanToken:
		return "<"
	case ast.KindGreaterThanToken:
		return ">"
	case ast.KindLessThanEqualsToken:
		return "<="
	case ast.KindGreaterThanEqualsToken:
		return ">="
	case ast.KindAmpersandAmpersandToken:
		return "&&"
	case ast.KindBarBarToken:
		return "||"
	default:
		return kind.String()
	}
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

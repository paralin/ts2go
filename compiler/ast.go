package compiler

// ASTNode represents a node in the TypeScript AST.
type ASTNode struct {
	Kind       string     `json:"kind"`
	KindNumber int        `json:"kindNumber"`
	Pos        int        `json:"pos"`
	End        int        `json:"end"`
	Text       string     `json:"text"`
	Name       string     `json:"name,omitempty"`
	Value      string     `json:"value,omitempty"`
	Children   []ASTNode  `json:"children,omitempty"`
}

// TypeScript SyntaxKind constants (partial list of common kinds)
const (
	// Literals
	SyntaxKindNumericLiteral      = 8
	SyntaxKindStringLiteral       = 10
	SyntaxKindTrueKeyword         = 110
	SyntaxKindFalseKeyword        = 95
	
	// Identifiers and Names
	SyntaxKindIdentifier          = 79
	
	// Declarations
	SyntaxKindVariableStatement   = 237
	SyntaxKindVariableDeclaration = 254
	SyntaxKindFunctionDeclaration = 256
	SyntaxKindParameter           = 165
	
	// Types
	SyntaxKindNumberKeyword       = 148
	SyntaxKindStringKeyword       = 152
	SyntaxKindBooleanKeyword      = 136
	SyntaxKindVoidKeyword         = 116
	
	// Expressions
	SyntaxKindBinaryExpression    = 220
	SyntaxKindCallExpression      = 207
	SyntaxKindPropertyAccessExpression = 205
	
	// Statements
	SyntaxKindReturnStatement     = 245
	SyntaxKindExpressionStatement = 238
	SyntaxKindBlock               = 233
	SyntaxKindIfStatement         = 239
	
	// Other
	SyntaxKindSourceFile          = 305
	SyntaxKindEndOfFileToken      = 1
)

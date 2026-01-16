#!/usr/bin/env node
/**
 * TypeScript Parser Wrapper
 * 
 * This script uses the TypeScript compiler API to parse TypeScript source code
 * and outputs the AST in JSON format that can be consumed by the Go transpiler.
 * 
 * Usage: node parser.js <input-file>
 */

import * as ts from 'typescript';
import * as fs from 'fs';

function parseTypeScriptFile(filename) {
  const sourceCode = fs.readFileSync(filename, 'utf8');
  
  // Create a source file
  const sourceFile = ts.createSourceFile(
    filename,
    sourceCode,
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TS
  );

  // Convert AST to a serializable format
  function serializeNode(node) {
    const obj = {
      kind: ts.SyntaxKind[node.kind],
      kindNumber: node.kind,
      pos: node.pos,
      end: node.end,
      text: node.getText(sourceFile),
    };

    // Add specific properties based on node type
    if (ts.isIdentifier(node)) {
      obj.name = node.text;
    } else if (ts.isStringLiteral(node) || ts.isNumericLiteral(node)) {
      obj.value = node.text;
    } else if (ts.isFunctionDeclaration(node) || ts.isMethodDeclaration(node)) {
      if (node.name) {
        obj.name = node.name.getText(sourceFile);
      }
    } else if (ts.isVariableDeclaration(node)) {
      obj.name = node.name.getText(sourceFile);
    } else if (ts.isClassDeclaration(node) || ts.isInterfaceDeclaration(node)) {
      if (node.name) {
        obj.name = node.name.getText(sourceFile);
      }
    }

    // Recursively serialize children
    const children = [];
    ts.forEachChild(node, (child) => {
      children.push(serializeNode(child));
    });
    
    if (children.length > 0) {
      obj.children = children;
    }

    return obj;
  }

  const ast = serializeNode(sourceFile);
  return JSON.stringify(ast, null, 2);
}

// Main execution
const args = process.argv.slice(2);
if (args.length === 0) {
  console.error('Usage: node parser.js <input-file>');
  process.exit(1);
}

const inputFile = args[0];
try {
  const result = parseTypeScriptFile(inputFile);
  console.log(result);
} catch (error) {
  console.error('Error parsing TypeScript file:', error.message);
  process.exit(1);
}

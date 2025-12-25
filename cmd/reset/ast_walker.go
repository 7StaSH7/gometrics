package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"strings"
)

const generateComment = "generate:reset"

// findStructsWithComment scans a file's AST for struct declarations
// that have the "// generate:reset" comment.
func findStructsWithComment(file *ast.File, typesInfo *types.Info) []*structInfo {
	var structs []*structInfo

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		if genDecl.Doc == nil {
			continue
		}

		hasGenerateComment := false
		for _, comment := range genDecl.Doc.List {
			if strings.Contains(comment.Text, generateComment) {
				hasGenerateComment = true
				break
			}
		}

		if !hasGenerateComment {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			si := &structInfo{
				Name:   typeSpec.Name.Name,
				Fields: extractFields(structType, typesInfo),
			}
			structs = append(structs, si)
		}
	}

	return structs
}

// extractFields extracts information about exported fields from a struct type.
func extractFields(structType *ast.StructType, typesInfo *types.Info) []*fieldInfo {
	var fields []*fieldInfo

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}

		for _, name := range field.Names {
			if !name.IsExported() {
				continue
			}

			fi := analyzeFieldType(name.Name, field.Type, typesInfo)
			fields = append(fields, fi)
		}
	}

	return fields
}

// analyzeFieldType creates a fieldInfo by analyzing the field's type.
func analyzeFieldType(name string, expr ast.Expr, typesInfo *types.Info) *fieldInfo {
	fi := &fieldInfo{
		Name:     name,
		TypeExpr: exprToString(expr),
	}

	if typesInfo != nil && typesInfo.TypeOf(expr) != nil {
		typeInfo := typesInfo.TypeOf(expr)
		analyzeType(fi, typeInfo)
	} else {
		analyzeTypeFromAST(fi, expr)
	}

	return fi
}

// exprToString converts an AST expression to its string representation.
func exprToString(expr ast.Expr) string {
	var buf bytes.Buffer
	format.Node(&buf, token.NewFileSet(), expr)
	return buf.String()
}

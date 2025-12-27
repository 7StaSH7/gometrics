package main

import (
	"go/ast"
	"go/types"
	"strings"
)

// analyzeType populates fieldInfo based on Go type information.
func analyzeType(fi *fieldInfo, t types.Type) {
	switch typ := t.(type) {
	case *types.Basic:
		fi.TypeKind = typ.Kind()
	case *types.Slice:
		fi.IsSlice = true
		fi.ElemType = typ.Elem().String()
	case *types.Map:
		fi.IsMap = true
	case *types.Pointer:
		fi.IsPtr = true
		fi.ElemType = typ.Elem().String()
		if _, ok := typ.Elem().Underlying().(*types.Struct); ok {
			fi.IsStruct = true
		}
	case *types.Struct:
		fi.IsStruct = true
	case *types.Named:
		if _, ok := typ.Underlying().(*types.Struct); ok {
			fi.IsStruct = true
		} else {
			analyzeType(fi, typ.Underlying())
		}
	}
}

// analyzeTypeFromAST populates fieldInfo based on AST expression analysis.
func analyzeTypeFromAST(fi *fieldInfo, expr ast.Expr) {
	switch t := expr.(type) {
	case *ast.ArrayType:
		if t.Len == nil {
			fi.IsSlice = true
			fi.ElemType = exprToString(t.Elt)
		}
	case *ast.MapType:
		fi.IsMap = true
	case *ast.StarExpr:
		fi.IsPtr = true
		fi.ElemType = exprToString(t.X)
		if _, ok := t.X.(*ast.StructType); ok {
			fi.IsStruct = true
		} else if ident, ok := t.X.(*ast.Ident); ok {
			if ident.Name != "" && ident.Name[0] >= 'A' && ident.Name[0] <= 'Z' {
				fi.IsStruct = true
			}
		}
	case *ast.Ident:
		switch t.Name {
		case "bool", "string", "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
			"byte", "rune", "float32", "float64", "complex64", "complex128":
		default:
			if t.Name[0] >= 'A' && t.Name[0] <= 'Z' {
				fi.IsStruct = true
			}
		}
	case *ast.StructType:
		fi.IsStruct = true
	}
}

// getBaseTypeName extracts the base type name from a qualified type name.
func getBaseTypeName(typeName string) string {
	parts := strings.Split(typeName, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return typeName
}

// getZeroValue returns the zero value representation for a given type name.
func getZeroValue(typeName string) string {
	typeName = strings.TrimPrefix(typeName, "*")

	switch typeName {
	case "bool":
		return "false"
	case "string":
		return `""`
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"byte", "rune", "float32", "float64", "complex64", "complex128":
		return "0"
	default:
		baseType := getBaseTypeName(typeName)
		return baseType + "{}"
	}
}

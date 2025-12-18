// Package main implements a code generator that creates Reset() methods for structs.
//
// The generator scans the codebase for structs marked with the "// generate:reset" comment
// and generates Reset() methods that reset all fields to their zero values.
//
// Usage:
//
//	go run ./cmd/reset/              # scan current directory
//	go run ./cmd/reset/ -dir <path>  # scan specified directory
//
// The generator creates a reset.gen.go file in each package containing structs
// marked with the generate:reset comment.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	// generateComment is the magic comment that marks structs for Reset() generation
	generateComment = "generate:reset"
	// outputFileName is the name of the generated file
	outputFileName = "reset.gen.go"
)

// structInfo holds information about a struct that needs Reset() method.
type structInfo struct {
	Name   string       // Name of the struct
	Fields []*fieldInfo // List of struct fields
}

// fieldInfo holds information about a struct field.
type fieldInfo struct {
	Name     string          // Field name
	TypeExpr string          // Type expression as string
	TypeKind types.BasicKind // Kind for basic types
	IsSlice  bool            // True if field is a slice
	IsMap    bool            // True if field is a map
	IsPtr    bool            // True if field is a pointer
	IsStruct bool            // True if field is a struct
	ElemType string          // Element type for pointers and slices
}

func main() {
	var rootDir string
	flag.StringVar(&rootDir, "dir", ".", "root directory to scan")
	flag.Parse()

	if err := run(rootDir); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// run is the main entry point that loads all packages and processes them.
// It scans all packages recursively from the root directory.
func run(rootDir string) error {
	// Configure package loading with all necessary information
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: rootDir,
	}

	// Load all packages recursively
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	// Check for errors in loaded packages
	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("packages contain errors")
	}

	// Process each package
	for _, pkg := range pkgs {
		if err := processPackage(pkg); err != nil {
			return fmt.Errorf("failed to process package %s: %w", pkg.PkgPath, err)
		}
	}

	return nil
}

// processPackage scans a single package for structs with generate:reset comment
// and generates the reset.gen.go file if any are found.
func processPackage(pkg *packages.Package) error {
	structs := make([]*structInfo, 0)

	// Scan all files in the package
	for i, file := range pkg.Syntax {
		fileStructs := findStructsWithComment(file, pkg.TypesInfo)
		if len(fileStructs) > 0 {
			fileName := "unknown"
			if i < len(pkg.GoFiles) {
				fileName = pkg.GoFiles[i]
			} else if i < len(pkg.CompiledGoFiles) {
				fileName = pkg.CompiledGoFiles[i]
			}
			log.Printf("Found %d struct(s) in %s", len(fileStructs), fileName)
			structs = append(structs, fileStructs...)
		}
	}

	// Skip if no structs found with generate:reset comment
	if len(structs) == 0 {
		return nil
	}

	// Generate reset.gen.go file for this package
	if err := generateResetFile(pkg, structs); err != nil {
		return fmt.Errorf("failed to generate reset file: %w", err)
	}

	return nil
}

// findStructsWithComment scans a file's AST for struct declarations
// that have the "// generate:reset" comment.
func findStructsWithComment(file *ast.File, typesInfo *types.Info) []*structInfo {
	var structs []*structInfo

	// Iterate over all declarations in the file
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		// Check if declaration has documentation
		if genDecl.Doc == nil {
			continue
		}

		// Check for generate:reset comment
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

		// Process each type spec in the declaration
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Extract struct information
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
// Embedded fields are skipped, and only exported fields are included.
func extractFields(structType *ast.StructType, typesInfo *types.Info) []*fieldInfo {
	var fields []*fieldInfo

	for _, field := range structType.Fields.List {
		// Skip embedded fields (no names)
		if len(field.Names) == 0 {
			continue
		}

		// Process each named field
		for _, name := range field.Names {
			// Skip unexported fields
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
// It uses type information if available, otherwise falls back to AST analysis.
func analyzeFieldType(name string, expr ast.Expr, typesInfo *types.Info) *fieldInfo {
	fi := &fieldInfo{
		Name:     name,
		TypeExpr: exprToString(expr),
	}

	// Try to use type information first
	if typesInfo != nil && typesInfo.TypeOf(expr) != nil {
		typeInfo := typesInfo.TypeOf(expr)
		analyzeType(fi, typeInfo)
	} else {
		// Fallback to AST-based analysis
		analyzeTypeFromAST(fi, expr)
	}

	return fi
}

// analyzeType populates fieldInfo based on Go type information.
// It determines whether the field is a slice, map, pointer, struct, or basic type.
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
		// Check if pointer points to a struct
		if _, ok := typ.Elem().Underlying().(*types.Struct); ok {
			fi.IsStruct = true
		}
	case *types.Struct:
		fi.IsStruct = true
	case *types.Named:
		// For named types, check the underlying type
		if _, ok := typ.Underlying().(*types.Struct); ok {
			fi.IsStruct = true
		} else {
			analyzeType(fi, typ.Underlying())
		}
	}
}

// analyzeTypeFromAST populates fieldInfo based on AST expression analysis.
// This is a fallback when type information is not available.
func analyzeTypeFromAST(fi *fieldInfo, expr ast.Expr) {
	switch t := expr.(type) {
	case *ast.ArrayType:
		// If Len is nil, it's a slice
		if t.Len == nil {
			fi.IsSlice = true
			fi.ElemType = exprToString(t.Elt)
		}
	case *ast.MapType:
		fi.IsMap = true
	case *ast.StarExpr:
		fi.IsPtr = true
		fi.ElemType = exprToString(t.X)
		// Check if it's a pointer to struct
		if _, ok := t.X.(*ast.StructType); ok {
			fi.IsStruct = true
		} else if ident, ok := t.X.(*ast.Ident); ok {
			// Assume exported identifiers are struct types
			if ident.Name != "" && ident.Name[0] >= 'A' && ident.Name[0] <= 'Z' {
				fi.IsStruct = true
			}
		}
	case *ast.Ident:
		// Check if it's a basic type
		switch t.Name {
		case "bool", "string", "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
			"byte", "rune", "float32", "float64", "complex64", "complex128":
			// It's a basic type, no special handling needed
		default:
			// Assume exported identifiers are struct types
			if t.Name[0] >= 'A' && t.Name[0] <= 'Z' {
				fi.IsStruct = true
			}
		}
	case *ast.StructType:
		fi.IsStruct = true
	}
}

// exprToString converts an AST expression to its string representation.
func exprToString(expr ast.Expr) string {
	var buf bytes.Buffer
	format.Node(&buf, token.NewFileSet(), expr)
	return buf.String()
}

// generateResetFile creates a reset.gen.go file containing Reset() methods
// for all structs in the given package.
func generateResetFile(pkg *packages.Package, structs []*structInfo) error {
	var buf bytes.Buffer

	// Write file header
	fmt.Fprintf(&buf, "// Code generated by cmd/reset; DO NOT EDIT.\n\n")
	fmt.Fprintf(&buf, "package %s\n\n", pkg.Name)

	// Generate Reset() method for each struct
	for _, s := range structs {
		generateResetMethod(&buf, s)
		buf.WriteString("\n")
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("Warning: failed to format generated code: %v", err)
		formatted = buf.Bytes()
	}

	// Determine output file path
	var outputPath string
	if len(pkg.GoFiles) > 0 {
		dir := filepath.Dir(pkg.GoFiles[0])
		outputPath = filepath.Join(dir, outputFileName)
	} else if len(pkg.CompiledGoFiles) > 0 {
		dir := filepath.Dir(pkg.CompiledGoFiles[0])
		outputPath = filepath.Join(dir, outputFileName)
	} else {
		return fmt.Errorf("no files in package %s", pkg.PkgPath)
	}

	// Write to file
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	log.Printf("Generated %s for package %s", outputPath, pkg.Name)
	return nil
}

// generateResetMethod generates a Reset() method for a single struct.
func generateResetMethod(buf *bytes.Buffer, s *structInfo) {
	fmt.Fprintf(buf, "// Reset resets all fields of %s to their zero values\n", s.Name)
	fmt.Fprintf(buf, "func (r *%s) Reset() {\n", s.Name)

	for _, field := range s.Fields {
		generateFieldReset(buf, field)
	}

	fmt.Fprintf(buf, "}")
}

// generateFieldReset generates the reset code for a single field.
// The reset behavior depends on the field type:
// - Slices: truncate to zero length, preserve capacity
// - Maps: clear all entries
// - Pointers: reset pointed value if not nil
// - Structs: call Reset() if available, otherwise set to zero value
// - Primitives: set to zero value
func generateFieldReset(buf *bytes.Buffer, field *fieldInfo) {
	fieldRef := fmt.Sprintf("r.%s", field.Name)

	switch {
	case field.IsSlice:
		// Truncate slice but preserve capacity
		fmt.Fprintf(buf, "\t%s = %s[:0]\n", fieldRef, fieldRef)

	case field.IsMap:
		// Clear map using built-in clear function
		fmt.Fprintf(buf, "\tclear(%s)\n", fieldRef)

	case field.IsPtr:
		if field.IsStruct {
			// Pointer to struct: try to call Reset(), otherwise set to zero value
			fmt.Fprintf(buf, "\tif %s != nil {\n", fieldRef)
			fmt.Fprintf(buf, "\t\tif resetter, ok := any(%s).(interface{ Reset() }); ok {\n", fieldRef)
			fmt.Fprintf(buf, "\t\t\tresetter.Reset()\n")
			fmt.Fprintf(buf, "\t\t} else {\n")
			fmt.Fprintf(buf, "\t\t\t*%s = %s{}\n", fieldRef, getBaseTypeName(field.ElemType))
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t}\n")
		} else {
			// Pointer to primitive: reset to zero value
			fmt.Fprintf(buf, "\tif %s != nil {\n", fieldRef)
			zeroValue := getZeroValue(field.ElemType)
			fmt.Fprintf(buf, "\t\t*%s = %s\n", fieldRef, zeroValue)
			fmt.Fprintf(buf, "\t}\n")
		}

	case field.IsStruct:
		// Struct: try to call Reset(), otherwise set to zero value
		fmt.Fprintf(buf, "\tif resetter, ok := any(&%s).(interface{ Reset() }); ok {\n", fieldRef)
		fmt.Fprintf(buf, "\t\tresetter.Reset()\n")
		fmt.Fprintf(buf, "\t} else {\n")
		fmt.Fprintf(buf, "\t\t%s = %s{}\n", fieldRef, field.TypeExpr)
		fmt.Fprintf(buf, "\t}\n")

	default:
		// Primitive type: set to zero value
		zeroValue := getZeroValue(field.TypeExpr)
		fmt.Fprintf(buf, "\t%s = %s\n", fieldRef, zeroValue)
	}
}

// getBaseTypeName extracts the base type name from a qualified type name.
// For example, "pkg.Type" becomes "Type".
func getBaseTypeName(typeName string) string {
	parts := strings.Split(typeName, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return typeName
}

// getZeroValue returns the zero value representation for a given type name.
// For primitives, it returns the literal zero value (e.g., "0", "false", "").
// For other types, it returns the zero value constructor (e.g., "TypeName{}").
func getZeroValue(typeName string) string {
	// Remove pointer prefix if present
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
		// For named types, return zero value constructor
		baseType := getBaseTypeName(typeName)
		return baseType + "{}"
	}
}

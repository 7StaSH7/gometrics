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
	"flag"
	"fmt"
	"go/types"
	"log"

	"golang.org/x/tools/go/packages"
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
func run(rootDir string) error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Dir: rootDir,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("packages contain errors")
	}

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

	if len(structs) == 0 {
		return nil
	}

	if err := generateResetFile(pkg, structs); err != nil {
		return fmt.Errorf("failed to generate reset file: %w", err)
	}

	return nil
}

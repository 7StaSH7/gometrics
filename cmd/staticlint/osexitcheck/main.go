// Package osexitcheck implements a custom analyzer that checks for direct os.Exit calls
// in the main function of the main package.
package osexitcheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "check for direct os.Exit calls in main function of main package",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if fn.Name.Name != "main" {
				return true
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				selExpr, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "os" && selExpr.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "direct call to os.Exit is not allowed in main function")
				}
				return true
			})

			return false
		})
	}
	return nil, nil
}

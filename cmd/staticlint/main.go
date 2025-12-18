// staticlint is a custom static analysis tool that combines multiple analyzers.
//
// Usage:
//
//	staticlint [flags] packages
//
// Examples:
//
//	staticlint ./...
//	staticlint -json ./...
//
// Analyzers included:
//
// Standard analyzers (golang.org/x/tools/go/analysis/passes):
//   - printf: Checks consistency of Printf format strings and arguments.
//   - shift: Checks for shifts that equal or exceed the width of the integer.
//   - structtag: Checks that struct field tags conform to reflect.StructTag.Canonical.
//
// Staticcheck.io analyzers:
//   - SA*: All analyzers from the SA class (static checks).
//   - S*: All analyzers from the S class (simple code simplifications).
//   - ST1000: Checks that package comments are present and formatted correctly.
//
// Public analyzers:
//   - bodyclose: Checks whether HTTP response bodies are closed successfully.
//   - nilerr: Checks that there is no nil error getting returned in a specific case.
//
// Custom analyzers:
//   - osexitcheck: Checks for direct os.Exit calls in main function of main package.
package main

import (
	"strings"

	"github.com/7StaSH7/gometrics/cmd/staticlint/osexitcheck"
	"github.com/gostaticanalysis/nilerr"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	var mychecks []*analysis.Analyzer

	// 1. Standard static analyzers from golang.org/x/tools/go/analysis/passes
	mychecks = append(mychecks, printf.Analyzer)
	mychecks = append(mychecks, shift.Analyzer)
	mychecks = append(mychecks, structtag.Analyzer)

	// 2. All analyzers of class SA from staticcheck.io
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	// 3. One analyzer of other classes from staticcheck.io
	for _, v := range simple.Analyzers {
		mychecks = append(mychecks, v.Analyzer)
	}

	// One Style check (ST class)
	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1000" {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	// 4. Two public analyzers
	mychecks = append(mychecks, bodyclose.Analyzer)
	mychecks = append(mychecks, nilerr.Analyzer)

	// 5. Custom analyzer
	mychecks = append(mychecks, osexitcheck.Analyzer)

	multichecker.Main(mychecks...)
}
